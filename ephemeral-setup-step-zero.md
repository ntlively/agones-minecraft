sudo apt-get install kubectl
sudo apt-get update && sudo apt-get install -y gnupg software-properties-common

wget -O- https://apt.releases.hashicorp.com/gpg | \
gpg --dearmor | \
sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg

gpg --no-default-keyring \
--keyring /usr/share/keyrings/hashicorp-archive-keyring.gpg \
--fingerprint

echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
sudo tee /etc/apt/sources.list.d/hashicorp.list

sudo apt update

sudo apt-get install terraform

//had to run everything from local gcloud terminal, not the online one.
//managed dns zone name is duckzone and project id is agones-minecraft-384717 

//had to import my google domain into gcloud
gcloud beta domains registrations list-importable-domains
gcloud domains registrations import duckmaster.games

//had to manually install and create agones namespace before applying the tf version
kubectl create namespace agones-system
kubectl apply --server-side -f https://raw.githubusercontent.com/googleforgames/agones/release-1.31.0/install/yaml/install.yaml
kubectl apply -f k8s/mc-server-fleet.yml

//had to update to use gcloud dns  https://cloud.google.com/kubernetes-engine/docs/how-to/cloud-dns#gcloud_1
gcloud container clusters update minecraft \
    --cluster-dns=clouddns \
    --cluster-dns-scope=cluster \
    --region=us-east1-b
//then the pod pool needs the same
gcloud container clusters upgrade minecraft \
    --node-pool=default \
    --region=us-east1-b
