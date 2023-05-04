package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/go-co-op/gocron"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/kardianos/service"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var (
	dbname          string = "db.sqlite"
	executable_path string
	app             fiber.App
	logger          service.Logger
	logrus_logger   logrus.Logger
	log_path        string
	port            string = "3000"
)

const ENV int = 0 // 0:Dev, 1: Prod

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if ENV == 1 {
		executable_path = filepath.Dir(ex)
		dbname = filepath.Join(executable_path, "db.sqlite")
		log_path = filepath.Join(executable_path, "log.txt")
	} else {
		executable_path = ""
		dbname = "db.sqlite"
		log_path = "log.txt"
		port = "3001"
	}

	f, err := os.OpenFile(log_path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	logrus_logger = logrus.Logger{
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}

	logrus_logger.Printf("DB: %s", dbname)

	serviceConfig := &service.Config{
		Name:        "ServiceManager",
		DisplayName: "Service Manager",
		Description: "Service to Restart Managed Services",
	}

	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		logrus_logger.Errorf("Cannot Create Service %s", err.Error())
	}

	logger, err = s.Logger(nil)
	if err != nil {
		logrus_logger.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		logrus_logger.Errorf("Cannot Start Service %s", err.Error())
	}
}

/**
Service Code Configuration
*/

type program struct{}

func (p program) Start(s service.Service) error {
	logrus_logger.Println(s.String() + " started!")
	go p.run()

	return nil
}

func (p program) Stop(s service.Service) error {
	logrus_logger.Info(s.String() + " stopped!")
	app.Shutdown()
	return nil
}

func (p program) run() {
	logrus_logger.Info("Starting Services!")
	logrus_logger.Info(runtime.GOOS)

	scheduleServices()

	startWebServer()

}

/*
*
 */
type CronJob struct {
	gorm.Model
	Name        string `json:"name"`
	ServiceName string `json:"service_name"`
	Schedule    string `json:"schedule"`
}

func listJobs() []CronJob {
	db, err := DBConnect(dbname)
	if err != nil {
		panic("Failed to connect to Data Base!")
	}
	db.AutoMigrate(&CronJob{})

	var jobs []CronJob
	db.Find(&jobs)

	return jobs
}

func UseCronJobRouter(router fiber.Router) {
	router.Get("api/", func(c *fiber.Ctx) error {
		logrus_logger.Info("Api")
		return c.SendString("Go Application API!")
	})

	router.Get("api/jobs", func(c *fiber.Ctx) error {
		jobs := listJobs()
		// return c.JSON(fiber.Map{
		// 	"jobs": jobs,
		// })
		return c.JSON(jobs)
	})

	router.Post("api/job", func(c *fiber.Ctx) error {
		job := CronJob{}
		if err := c.BodyParser(&job); err != nil {
			return err
		}
		db, err := DBConnect(dbname)
		if err != nil {
			panic("Failed to connect ")
		}
		db.AutoMigrate(&CronJob{})

		result := db.Create(&job)
		if result.Error != nil {
			logrus_logger.Error(result.Error)
		}
		logrus_logger.Info(job.ID)
		new_job := CronJob{}
		db.First(&new_job, job.ID)
		scheduleServiceAt(new_job)
		return c.Status(fiber.StatusCreated).JSON(new_job)
	})

	router.Put("api/job/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err,
			})
		}
		job := CronJob{}
		if err := c.BodyParser(&job); err != nil {
			return err
		}
		db, err := DBConnect(dbname)
		if err != nil {
			panic("Failed to connect ")
		}

		job_lkp := CronJob{}
		db.First(&job_lkp, id)
		old_schedule := job_lkp.Schedule
		job_lkp.Name = job.Name
		job_lkp.ServiceName = job.ServiceName
		job_lkp.Schedule = job.Schedule
		db.Save(&job_lkp)
		logrus_logger.Infof("Job Schedule %v is Updated!", job_lkp.ID)
		if job.Schedule != old_schedule {
			scheduleServiceAt(job_lkp)
			logrus_logger.Infof("Job Re-Schedule %v !", job_lkp.ServiceName)
		}
		return c.Status(fiber.StatusOK).JSON(job_lkp)
	})

	router.Patch("api/job/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err,
			})
		}

		db, err := DBConnect(dbname)
		if err != nil {
			panic("Failed to connect ")
		}
		job := CronJob{}
		db.First(&job, id)
		logrus_logger.Infof("Job Schedule %v is Restarted!", job.ID)
		//go restartService(job.ServiceName)
		go resetNow(job)
		return c.Status(fiber.StatusOK).JSON(job)
	})

	router.Patch("api/job/stop/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err,
			})
		}

		db, err := DBConnect(dbname)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err,
			})
		}
		job := CronJob{}
		db.First(&job, id)
		go stopService(job)
		return c.Status(fiber.StatusOK).JSON(job)
	})

}

/**
 */
func DBConnect(dbname string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		panic("Failed to connect ")
	}

	return db, err
}

/*
*
 */
func startWebServer() {

	app := fiber.New(fiber.Config{
		DisableStartupMessage: ENV == 1,
	})

	// CORS
	app.Use(cors.New(cors.Config{
		AllowHeaders: "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins: "*",
		//AllowCredentials: true,
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	UseCronJobRouter(app)

	public_path := filepath.Join(executable_path, "public")
	app.Static("/", public_path, fiber.Static{Index: "index.html"})
	//app.Static("/", "./public", fiber.Static{Index: "index.html"})

	logrus_logger.Infof("Starting Web Server on http://localhost:%s", port)
	err := app.Listen(":" + port)
	if err != nil {
		panic(err)
	}
}

/*
*
Schedule Services
*/
func scheduleServices() {
	jobs := listJobs()
	logrus_logger.Infof("Scheduling Services %v.....", len(jobs))
	for _, schedule := range jobs {
		scheduleServiceAt(schedule)
	}
}
func scheduleServiceAt(job CronJob) {
	logrus_logger.Infof("Scheduling %s, Service Name %s, At %s", job.Name, job.ServiceName, job.Schedule)
	t := time.Now()
	schedule := gocron.NewScheduler(t.Location())
	schedule.Every(1).Day().At(job.Schedule).Do(restartService, job.ServiceName)
	//msolap.Cron("*/1 9 * * *").Do(restartService, serviceName)
	schedule.StartAsync()
}

func resetNow(job CronJob) {
	// channel := make(chan string)
	// go restartService(job.ServiceName, channel)

	logrus_logger.Infof("Scheduling Reset Now %s, Service Name %s, At %s", job.Name, job.ServiceName, job.Schedule)
	// t := time.Now()
	// schedule := gocron.NewScheduler(t.Location())
	// schedule.StartAt(time.Now().Add(time.Second*10)).Do(restartService, job.ServiceName)
	// schedule.StartAsync()
	restartService(job.ServiceName)
}

func restartService(name string) error {
	logrus_logger.Infof("Executing Schedule Task %s", name)
	m, err := mgr.Connect()
	if err != nil {
		logrus_logger.Error(err)
		return nil
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		logrus_logger.Info(err)
		return nil
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		logrus_logger.Error(err)
		return nil
	}
	if status.State != svc.Stopped {
		logrus_logger.Infof("Stoping Service %s", name)
		_, err = s.Control(svc.Stop)
		if err != nil {
			logrus_logger.Error(err)
			return nil
		}
	}
	time_sleep := 10
	counter := 0
	status, _ = s.Query()

	for status.State != svc.Stopped {
		counter = counter + 1
		time.Sleep(time.Duration(time_sleep) * time.Second)
		logrus_logger.Infof("Waiting for Stop Service %s, Seconds %v, State %v", name, counter*time_sleep, status.State)
		if status.State == svc.Stopped {
			logrus_logger.Info("Service Stopped!")
			break
		}
		status, _ = s.Query()
	}

	logrus_logger.Infof("Trying Start Service...... %s!", name)

	err = s.Start("is", "manual-started")

	counter = 0
	status, _ = s.Query()
	for status.State != svc.Running {
		counter = counter + 1
		time.Sleep(time.Duration(time_sleep) * time.Second)
		logrus_logger.Infof("Waiting for Start Service %s, Seconds %v, State %v", name, counter*time_sleep, status.State)
		if status.State == svc.Running {
			logrus_logger.Info("Service Running!")
			break
		}
		status, _ = s.Query()
	}

	if err != nil {
		logrus_logger.Error(err)
		return nil
	}

	logrus_logger.Infof("Servicio %s reiniciado", name)
	return nil
}

func stopService(job CronJob) error {
	logrus_logger.Infof("Stopping Service %s", job.ServiceName)
	m, err := mgr.Connect()
	if err != nil {
		logrus_logger.Error(err)
		return nil
	}
	defer m.Disconnect()

	s, err := m.OpenService(job.ServiceName)
	if err != nil {
		logrus_logger.Info(err)
		return nil
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		logrus_logger.Error(err)
		return nil
	}
	if status.State != svc.Stopped {
		logrus_logger.Infof("Stoping Service %s", job.ServiceName)
		_, err = s.Control(svc.Stop)
		if err != nil {
			logrus_logger.Error(err)
			return nil
		}
	} else {
		logrus_logger.Infof("Service %s is yet Stopped!", job.ServiceName)
		return nil
	}
	time_sleep := 10
	counter := 0
	status, _ = s.Query()

	for status.State != svc.Stopped {
		counter = counter + 1
		time.Sleep(time.Duration(time_sleep) * time.Second)
		logrus_logger.Infof("Waiting for Stop Service %s, Seconds %v, State %v", job.Name, counter*time_sleep, status.State)
		if status.State == svc.Stopped {
			break
		}
		status, _ = s.Query()
	}
	logrus_logger.Infof("Service %s Stopped", job.ServiceName)

	return nil
}
