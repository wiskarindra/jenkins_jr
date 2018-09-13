# Spyro
<img width="460" alt="screen shot 2018-03-29 at 7 25 29 pm" src="https://user-images.githubusercontent.com/9160614/38088731-0f21b3e0-3387-11e8-9faf-4e601b648d5e.png">

## Description

**Spyro** (**Inspiration** :arrow_right: **Inspira** :arrow_right: **Spira** :arrow_right: **Spiro** :arrow_right: **Spyro**:heavy_exclamation_mark:) adalah _microservice_ untuk melayani fitur halaman inspirasi Bukalapak. Sebagai permulaan microservice ini dirancang untuk menampilkan daftar postingan inspirasi yang diunggah oleh admin Bukalapak. Setiap postingan disertai _tag_ yang memiliki _link_ untuk mengarahkan user ke pelapak ataupun produk-produk di Bukalapak.

## SLO and SLI
- Availability > 99%
- Mean Response Time < 80 ms

## Architecture Diagram

![jenkins_jr architecture diagram](https://user-images.githubusercontent.com/9160614/37889699-d6e1e8c4-30f7-11e8-84d4-9b9b29d000dd.png)

## Owner

WOW Squad

## Contact and On-Call Information

See [WOW Squad Members](https://bukalapak.atlassian.net/wiki/spaces/WOW/overview)

## Links

- [Confluence Page](https://bukalapak.atlassian.net/wiki/spaces/WOW/pages/485720567/Spyro)
- [Grafana Monitoring](https://grafana-mon.bldevs.info/dashboard/db/wow-microservices?orgId=1&from=now-1h&to=now&refresh=5s)

## Onboarding and Development Guide

### Prerequisite
- Git
- Go 1.9 or later
- MySQL 14.14 or later
- Docker
- Aleppo
- Go Dep

### Installation

- Install Git

  See [Git Installation](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

- Install Go (Golang)

  See [Golang Installation](https://golang.org/doc/install)

- Install MySQL

  See [MySQL Installation](https://www.mysql.com/downloads/)

- Install Docker

  See [Docker Installation](https://docs.docker.com/install/)

- Install Aleppo

  See [Aleppo Repository](https://github.com/bukalapak/aleppo)

- Install Go Dep

  See [Go Dep Installation](https://github.com/golang/dep#installation)

- Clone this repo in your local at `$GOPATH/src/github.com/bukalapak`

  If you have not set your GOPATH, set it using [this](https://golang.org/doc/code.html#GOPATH) guide.
  If you don't have directory `src`, `github.com`, or `bukalapak` in your GOPATH, please make them.

  ```sh
  git clone git@github.com:wiskarindra/jenkins_jr.git
  ```

- Go to Spyro directory, then sync the vendor file

  ```sh
  cd $GOPATH/src/github.com/wiskarindra/jenkins_jr
  make dep
  ```

- Copy env.sample and if necessary, modify the env value(s)

  ```sh
  cp env.sample .env
  ```

- Prepare database

  ```sh
  make setup
  ```

- Run Spyro

  ```sh
  go run app/jenkins_jr/main.go
  ```

- Check whether it is ran correctly. It should return `OK` message

  ```sh
  curl -X GET "http://localhost:7010/healthz"
  ```

## Request Flows, Endpoints, and Dependencies

### Request Flow

Stated in [Architecture Diagram](https://github.com/wiskarindra/jenkins_jr#architecture-diagram)

### Enpoints
[General](https://blueprint.bukalapak.io/#inspirations) | [Exclusive](https://blueprint.bukalapak.io/exclusive.html#inspirations)

### Dependencies
- [Aleppo](https://github.com/bukalapak/aleppo)
- [MySQL](https://www.mysql.com/)

## On-Call Runbooks

No alerting

## FAQ
Gagal nih pas jalanin `make setup`. Kenapa ya?
> Punya VPN buat akses datacenter Bukalapak? Coba nyalain dulu.

Masih gagal juga. Kenapa ya?
>Mac user? Coba ganti di `.env` nya bagian `DATABASE_HOST` sama `DATABASE_TEST_HOST` jadi `docker.for.mac.localhost`. Kalau masih gagal juga, atau bukan mac user coba hubungi salah satu dari [on-call person](https://github.com/wiskarindra/jenkins_jr#contact-and-on-call-information)
