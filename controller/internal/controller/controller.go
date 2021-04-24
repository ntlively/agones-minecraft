package controller

import (
	"context"
	"fmt"
	"strings"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	AnnotationPrefix      string = "agones-mc"
	HostnameAnnotation    string = "hostname"
	ExternalDnsAnnotation string = "externalDNS"
)

type GameServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	log    logr.Logger
	dns    provider.DnsClient
}

func (r *GameServerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	gs := &agonesv1.GameServer{}

	if err := r.Get(ctx, req.NamespacedName, gs); err != nil {
		if errors.IsNotFound(err) {
			r.log.Error(err, "Could not find GameServer")
			return reconcile.Result{}, nil
		}

		r.log.Error(err, "Error getting GameServer")
		return reconcile.Result{}, err
	}

	if ready := isServerAllocated(gs); !ready {
		r.log.Info("Waiting for port/address allocation")
		return reconcile.Result{}, nil
	}

	hostname, hostnameFound := getAnnotation(HostnameAnnotation, gs)
	_, externalDnsFound := getAnnotation(ExternalDnsAnnotation, gs)

	if externalDnsFound {
		r.log.Info(fmt.Sprintf("External DNS set for %s", gs.Name))
		return reconcile.Result{}, nil
	}

	if hostnameFound {
		_, err := r.dns.SetExternalDns(hostname, gs)

		if err != nil {

			switch err.(type) {
			case *provider.DNSRecordExists:
				r.log.Info(err.Error())
			default:
				r.log.Error(err, "Error creating DNS records")
				return reconcile.Result{}, nil
			}

		}

		r.Get(ctx, req.NamespacedName, gs)

		if err := setAnnotation(ExternalDnsAnnotation, mcDns.JoinARecordName(hostname, gs.Name), gs); err != nil {
			r.log.Error(err, "Error setting externalDNS annotation")
			return reconcile.Result{}, nil
		}

		if err := r.Update(ctx, gs); err != nil {
			return reconcile.Result{}, err
		}
		r.log.Info("GameServer updated")
	} else {
		r.log.Info(fmt.Sprintf("Not hostname annotation for %s", gs.Name))
	}

	return reconcile.Result{}, nil
}

func getAnnotation(suffix string, gs *agonesv1.GameServer) (string, bool) {
	annotation := fmt.Sprintf("%s/%s", AnnotationPrefix, suffix)
	hostname, ok := gs.Annotations[annotation]

	if !ok || strings.TrimSpace(hostname) == "" {
		return "", false
	}

	return hostname, true
}

func setAnnotation(suffix string, value string, gs *agonesv1.GameServer) error {
	annotation := fmt.Sprintf("%s/%s", AnnotationPrefix, ExternalDnsAnnotation)

	if value == "" || strings.TrimSpace(value) == "" {
		return fmt.Errorf("value for %s annotation empty", annotation)
	}

	gs.Annotations[annotation] = value
	return nil
}

func isServerAllocated(gs *agonesv1.GameServer) bool {
	if gs.Status.Address == "" || len(gs.Status.Ports) == 0 {
		return false
	}

	return true
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
	return &GameServerReconciler{client, scheme, log, dns}
}
