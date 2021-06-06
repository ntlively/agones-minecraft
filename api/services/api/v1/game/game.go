package game

import (
	"agones-minecraft/db"
	"errors"

	v1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/google/uuid"
	"gorm.io/gorm"

	gamev1Model "agones-minecraft/models/v1/game"
	"agones-minecraft/services/k8s/agones"
)

var (
	ErrSubdomainTaken     error = errors.New("custom subdomain not available")
	ErrGameServerNotFound error = errors.New("game server not found")
)

func GetGameById(game *gamev1Model.Game, ID uuid.UUID) error {
	game.ID = ID
	return db.DB().First(game).Error
}

func GetGameByName(game *gamev1Model.Game, name string) error {
	return db.DB().Where("name = ?", name).First(game).Error
}

func GetGameByUserIdAndName(game *gamev1Model.Game, userId uuid.UUID, name string) error {
	return db.DB().Where("user_id = ? AND name = ?", userId, name).First(game).Error
}

func CreateGame(game *gamev1Model.Game, gs *v1.GameServer) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		if game.CustomSubdomain != nil {
			if ok := agones.Client().HostnameAvailable(agones.GetDNSZone(), *game.CustomSubdomain); !ok {
				return ErrSubdomainTaken
			}
			agones.SetHostname(gs, agones.GetDNSZone(), *game.CustomSubdomain)
		}

		gameServer, err := agones.Client().CreateDryRun(gs)
		if err != nil {
			return err
		}

		// point to newly created gameserver obj
		*gs = *gameServer

		game.ID = uuid.MustParse(string(gs.UID))
		game.Name = gs.Name
		game.GameState = gamev1Model.On

		if err := db.DB().Create(game).Error; err != nil {
			return err
		}

		if _, err := agones.Client().Create(gs); err != nil {
			return err
		}

		return nil
	})
}

func DeleteGame(game *gamev1Model.Game, userId uuid.UUID, name string) error {
	return db.DB().Transaction(func(tx *gorm.DB) error {
		if err := GetGameByUserIdAndName(game, userId, name); err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrGameServerNotFound
			}
			return err
		}

		if err := db.DB().Delete(game).Error; err != nil {
			return err
		}

		if err := agones.Client().Delete(name); err != nil {
			return err
		}

		return nil
	})
}

func UpdateGame(game *gamev1Model.Game) error {
	return db.DB().Model(game).Updates(game).First(game).Error
}
