[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]

<br />
<p align="center">
  <h2 align="center">Agones Minecraft DNS Controller</h2>

  <p align="center">
  Custom Kubernetes controller for automating external DNS on Minecraft GameServers
    <br />
    <a href="https://github.com/saulmaldonado/agones-minecraft/tree/main/controller"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/saulmaldonado/agones-minecraft/tree/main/k8s/agones-mc-dns-controller">View Example</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-minecraft/issues">Report Bug</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-minecraft/issues">Request Feature</a>
  </p>
</p>

<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#acknowledgements">Acknowledgements</a></li>
    <li><a href="#author">Author</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

Custom Kubernetes controller for automating provisioning external DNS records for Minecraft GameServers using third-party cloud providers

### Built With

- [controller-runtime](https://github.com/Raqbit/mc-pinger)
- [Agones](agones.dev/agones)

<!-- GETTING STARTED -->

## Getting Started

To get the controller running:

```sh
docker run -it --rm saulmaldonado/agones-mc-dns-controller --gcp-project=<PROJECT_ID> --zone=<DNS_MANAGED_ZONE>
```

### Prerequisites

You need a running GKE cluster running with Agones resources and controllers installed

- GKE

```sh
gcloud container clusters create minecraft --cluster-version=1.18 \
  --tags=mc \
  --scopes=gke-default,"https://www.googleapis.com/auth/ndev.clouddns.readwrite" \
  --num-nodes=2 \
  --no-enable-autoupgrade \
  --machine-type=n2-standard-4
```

```sh
gcloud config set container/cluster minecraft
gcloud container clusters get-credentials minecraft
```

```sh
gcloud compute firewall-rules create mc-server-firewall \
  --allow tcp:7000-8000 \
  --target-tags mc \
  --description "Firewall rule to allow mc server tcp traffic"
```

- Agones

```sh
kubectl create namespace agones-system
kubectl apply -f https://raw.githubusercontent.com/googleforgames/agones/release-1.13.0/install/yaml/install.yaml
```

### Installation

#### Download from Docker Hub

```sh
docker pull saulmaldonado/agones-mc-dns-controller
```

<!-- USAGE EXAMPLES -->

## Usage

This controller takes advantage of dynamic host port allocation that Agones provisions for GameServers. Instead of creating external DNS records for Kubernetes services and ingresses, DNS `A` and `SRV` are created to point to GameServer host nodes and the their GameServer ports.

### Nodes

To provision an `A` record for Nodes, they need to contain a `agones-mc/hostname:` annotation that contains the domain for the current Managed Zone. This will indicate to the controller that the Node needs an `A` record.

This can be done using `kubectl`:

```sh
kubectl annotate node/<NODE_NAME> agones-mc/hostname=<DOMAIN>
```

The `A` record are generated with the name `<NODE_NAME>.<DOMAIN>.`
New annotation with `agones-mc/externalDNS` will contain the new `A` record that points to the Node IP.

### GameServer

To provision an `SRV` record for GameServers, they need to contain an `agones-mc/hostname` annotation that contains the domain for the current Managed Zone. This will indicate to the controller that the GameServer needs a `SRV` record.

#### GameServer Pod template example

```yml
template:
  metadata:
    annotations:
      agones-mc/hostname: <DOMAIN_NAME> # Domain name of the managed zone
      # agones-mc/externalDNS: <GAMESERVER_NAME>.<DOMAIN_NAME> # Will be added by the controller
  spec:
    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft server image
        imagePullPolicy: Always
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: 'TRUE'

      - name: mc-monitor
        image: saulmaldonado/agones-mc-monitor # Agones monitor sidecar
        imagePullPolicy: Always
```

Once the pod has been created, a new `SRV` will be generated with the format `_minecraft._tcp.<GAMESERVER_NAME>.<DOMAIN>.` that points to the `A` record of the host Node `0 0 <PORT> <HOST_A_RECORD>`.

A new annotation `agones-mc/externalDNS` will then be added to the GameServer containing the URL from which players can connect to.

For example: `mc-server-cfwd7.saulmaldonado.me` will connect to the Minecraft GameServer named `mc-server-cfwd7` on its dynamically allocated port.

#### [Full GameServer specification example](../k8s/mc-server.yml)

#### [Full Fleet specification example](../k8s/mc-server-fleet.yml)

### Run Locally with Docker

```sh
docker run -it --rm saulmaldonado/agones-mc-dns-controller --gcp-project=<PROJECT_ID> --zone=<DNS_MANAGED_ZONE> --kubeconfig=<KUBE_CONFIG_PATH>
```

Flags:

```
  --gcp-project string
        GCP project id
  --kubeconfig string
        Paths to a kubeconfig. Only required if out-of-cluster.
  --zone string
        DNS zone that the controller will manage
```

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/saulmaldonado/agones-minecraft/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Clone the Project
3. Create your Feature or Fix Branch (`git checkout -b (feat|fix)/AmazingFeatureOrFix`)
4. Commit your Changes (`git commit -m 'Add some AmazingFeatureOrFix'`)
5. Push to the Branch (`git push origin (feat|fix)/AmazingFeature`)
6. Open a Pull Request

### Build from source

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-minecraft.git
   ```

2. Build

   ```sh
   make go-build
   ```

### Build from Dockerfile

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-minecraft.git
   ```

2. Build

   ```sh
   docker build -t <hub-user>/agones-mc-dns-controller:latest ./controller
   ```

3. Push to Docker repo

   ```sh
   docker push <hub-user>/agones-mc-dns-controller:latest
   ```

<!-- LICENSE -->

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.

<!-- ACKNOWLEDGEMENTS -->

## Acknowledgements

- [itzg/docker-minecraft-server](https://github.com/itzg/docker-minecraft-server)

## Author

### Saul Maldonado

- 🐱 Github: [@saulmaldonado](https://github.com/saulmaldonado)
- 🤝 LinkedIn: [@saulmaldonado4](https://www.linkedin.com/in/saulmaldonado4/)
- 🐦 Twitter: [@saul_mal](https://twitter.com/saul_mal)
- 💻 Website: [saulmaldonado.com](https://saulmaldonado.com/)

## Show your support

Give a ⭐️ if this project helped you!

[contributors-shield]: https://img.shields.io/github/contributors/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[contributors-url]: https://github.com/saulmaldonado/agones-minecraft/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[forks-url]: https://github.com/saulmaldonado/agones-minecraft/network/members
[stars-shield]: https://img.shields.io/github/stars/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[stars-url]: https://github.com/saulmaldonado/agones-minecraft/stargazers
[issues-shield]: https://img.shields.io/github/issues/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[issues-url]: https://github.com/saulmaldonado/agones-minecraft/issues
[license-shield]: https://img.shields.io/github/license/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[license-url]: https://github.com/saulmaldonado/agones-minecraft/blob/master/LICENSE.txt