package cheek

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	election "github.com/joe-at-startupmedia/consul-leader-election"
	"gopkg.in/yaml.v3"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

// Schedule defines specs of a job schedule.
type Schedule struct {
	Jobs       map[string]*JobSpec `yaml:"jobs" json:"jobs"`
	OnSuccess  OnEvent             `yaml:"on_success,omitempty" json:"on_success,omitempty"`
	OnError    OnEvent             `yaml:"on_error,omitempty" json:"on_error,omitempty"`
	TZLocation string              `yaml:"tz_location,omitempty" json:"tz_location,omitempty"`
	loc        *time.Location
	log        zerolog.Logger
	cfg        Config
}

// Run a Schedule based on its specs.
func (s *Schedule) Run(e *election.Election) {
	var currentTickTime time.Time
	s.log.Info().Msg("Scheduler started")
	ticker := time.NewTicker(15 * time.Second) // could be longer
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			s.log.Debug().Msg("tick")

			if !e.IsLeader() {
				s.log.Debug().Msg("follower, doing nothing")
				continue
			}

			currentTickTime = s.now()

			for _, j := range s.Jobs {
				if j.Cron == "" {
					continue
				}

				if j.nextTick.Before(currentTickTime) {
					s.log.Debug().Msgf("%v is due", j.Name)
					// first set nextTick
					if err := j.setNextTick(currentTickTime, false); err != nil {
						s.log.Fatal().Err(err).Msg("error determining next tick")
					}

					go func(j *JobSpec) {
						j.execCommandWithRetry("cron")
					}(j)
				}
			}

		case sig := <-sigs:
			s.log.Info().Msgf("%s signal received, exiting...", sig.String())
			e.Stop()
			return
		}
	}
}

type stringArray []string

func (a *stringArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		*a = strings.Fields(single)
	} else {
		*a = multi
	}
	return nil
}

func readSpecs(fn string) (Schedule, error) {
	yfile, err := os.ReadFile(fn)
	if err != nil {
		return Schedule{}, err
	}

	specs := Schedule{}

	if err = yaml.Unmarshal(yfile, &specs); err != nil {
		return Schedule{}, err
	}

	return specs, nil
}

// initialize Schedule spec and logic.
func (s *Schedule) initialize() error {
	// validate tz location
	if s.TZLocation == "" {
		s.TZLocation = "Local"
	}

	loc, err := time.LoadLocation(s.TZLocation)
	if err != nil {
		return err
	}
	s.loc = loc

	for k, v := range s.Jobs {
		// check if trigger references exist
		triggerJobs := append(v.OnSuccess.TriggerJob, v.OnError.TriggerJob...)
		for _, t := range triggerJobs {
			if _, ok := s.Jobs[t]; !ok {
				return fmt.Errorf("cannot find spec of job '%s' that is referenced in job '%s'", t, k)
			}
		}
		// set some metadata & refs for each job
		// for easier retrievability
		v.Name = k
		v.globalSchedule = s
		v.log = s.log
		v.cfg = s.cfg

		// validate cron string
		if err := v.ValidateCron(); err != nil {
			return err
		}

		// init nextTick
		if err := v.setNextTick(s.now(), true); err != nil {
			return err
		}

	}

	return nil
}

func (s *Schedule) now() time.Time {
	return time.Now().In(s.loc)
}

func loadSchedule(log zerolog.Logger, cfg Config, fn string) (Schedule, error) {
	s, err := readSpecs(fn)
	if err != nil {
		return Schedule{}, err
	}
	s.log = log
	s.cfg = cfg

	// run validations
	if err := s.initialize(); err != nil {
		return Schedule{}, err
	}
	s.log.Info().Msg("Scheduled loaded and validated")
	return s, nil
}

// RunSchedule is the main entry entrypoint of cheek.
func RunSchedule(log zerolog.Logger, cfg Config, scheduleFn string) error {
	s, err := loadSchedule(log, cfg, scheduleFn)
	if err != nil {
		return err
	}
	numberJobs := len(s.Jobs)
	i := 1
	for k := range s.Jobs {
		s.log.Info().Msgf("Initializing (%v/%v) job: %s", i, numberJobs, k)
		i++
	}
	go server(&s)
	e := elector()
	s.Run(e)
	return nil
}

type notify struct {
	T string
}

func (n *notify) EventLeader(f bool) {
	if f {
		fmt.Println(n.T, "I'm the leader!")
	} else {
		fmt.Println(n.T, "I'm no longer the leader!")
	}
}

func elector() *election.Election {

	conf := api.DefaultConfig()
	consul, _ := api.NewClient(conf)
	n := &notify{
		T: "cheek-turner",
	}

	elconf := &election.ElectionConfig{
		CheckTimeout: 5 * time.Second,
		Client:       consul,
		Key:          "service/cheek-turner-election/leader",
		LogLevel:     election.LogDebug,
		Event:        n,
	}

	e := election.NewElection(elconf)

	go e.Init()

	return e
}
