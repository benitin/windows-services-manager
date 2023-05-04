# Windows Service Manager


**Windows Service Manager App** is a utility to **Schedule Restart**/**Restart**/**Stop** Windows Services.

* Scheduling Services to Restart.

![Scheduling Services to Restart](https://raw.githubusercontent.com/benitin/windows-services-manager/main/public/img/01-scheduling-service-restart.png "Scheduling Services to Restart")


* Managing Schedules

![Managing Schedules](https://raw.githubusercontent.com/benitin/windows-services-manager/main/public/img/02-managing-schedules.png "Managing Schedules")



### Create App

```shell
> go mod init monitoring.job
```

### Intalling Packages


[Fiber](https://gofiber.io/) An Express-inspired web framework writen in Go.

```shell
> go get github.com/gofiber/fiber/v2
```

[Logrus](https://github.com/sirupsen/logrus) is a structured logger for Go (golang), completely API compatible with the standart library logger.

```shell
> go get github.com/sirupsen/logrus
```

[Windows Services Manager](https://pkg.go.dev/golang.org/x/sys/windows/svc/mgr) package mgr can be used to manage Windows Services Programs. It can be used to install and remove then. It can use also start, stop and pause them. The package can query/change current service state and config parameters.

```shell
> go get golang.org/x/sys
```

[GORM](https://gorm.io/) Then fantastic ORM library for Golang
[Sqlite](https://gorm.io/docs/connecting_to_the_database.html) GORM officially supports the databases MySQL, PostgreSQL, SQLite, SQL Server and TiDB.

```shell
> go get -u gorm.io/gorm
> go get gorm.io/driver/sqlite
```

[GCRON](https://github.com/go-co-op/gocron),A Golang Job Scheduling Package, which lets you run Go functions at pre-determined intervals using a simple human-frienddly syntax.

```shell
> go get github.com/go-co-op/gocron
```

### Build Service Manager

[How to cross-compile Go programs for Windows, macOS, and Linux](https://freshman.tech/snippets/go/cross-compile-go-programs/)

* Linux
 ```shell
 > env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/job .
 ```
* Windows
```shell
> env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/service-manager.exe .
 ```


 ### Create Service for Windows

 [Crear servicios de Windows con SC [Service Controller]](https://www.zonasystem.com/2013/08/crear-servicios-de-windows-con-sc.html)

 ```shell
 > sc create ServiceManager binPath="path to executable" displayName="Service Manager" start=auto
 > sc description ServiceManager "Service to restart Managed Services"
 ```


 ```shell
 > sc delete ServiceManager 
 ```