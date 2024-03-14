
# cheek-turner

`cheek-turner` was forked from `cheek`: a pico-sized declarative job scheduler designed to excel in a single-node environment and aims to be lightweight, stand-alone and simple. It does not compete for robustness.

The following additions have been made:
1. Consul leader election to excel in a multi-node environment.

Note: `cheek-spreader` is not an upstream/downstream fork and has no-relation to this repository.

## Getting started

Everything about how you want the scheduler to function is defined in a schedule specification written in YAML. Start by creating this specification using the below example. Note, this structure should be more or less self-explanatory, if it is not, create an [issue](https://github.com/datarootsio/cheek/issues).

```yaml
tz_location: Europe/Brussels # optionally set timezone to adhere to
jobs:
  foo:
    command: date
    cron: "* * * * *" # a cron string to specify when to run
    on_success:
      trigger_job: # trigger something on run
        - bar
  bar:
    command: # command to run, use a list if you want to pass args
      - echo
      - $foo
    env: # you can pass env variables
      foo: bar
  other_workingdir:
    command: pwd
    working_directory: ../testdata # specify the working directory of the job
  coffee:
    command: this fails
    cron: "* * * * *"
    retries: 3
    on_error:
      notify_webhook: # notify something on error
        - https://webhook.site/4b732eb4-ba10-4a84-8f6b-30167b2f2762
      notify_slack_webhook: # notify slack via a slack compatible webhook
        - https://webhook.site/048ff47f-9ef5-43fb-9375-a795a8c5cbf5
```

If your `command` requires arguments, please make sure to pass them as an array like in `foo_job`.

Note that you can set `tz_location` if the system time of where you run your service is not to your liking.

## Scheduler

The core of `cheek` consists of a scheduler that uses the schedule specs defined in your `yaml` file to trigger jobs when they are due.

You can launch the scheduler via:

```sh
cheek run ./path/to/my-schedule.yaml
```

Check out `cheek run --help` for configuration options.

## Web UI

`cheek` ships with a web UI that by default gets launched on port `8081`. You can define the port on which it is accessible via the `--port` flag.

| ![main-screen](https://i.imgur.com/hq0Zxjb.png) |
| :---------------------------------------------: |
|                  main overview                  |

| ![detail](https://i.imgur.com/jc9wBQJ.png) |
| :----------------------------------------: |
|                 job detail                 |

You can access the UI by navigating to `http://localhost:8081`. When `cheek` is deployed you are recommended to NOT make this port publicly accessible, instead navigate to the UI via an SSH tunnel.

The UI allows to get a quick overview on jobs that have run, that error'd and their logs. It basically does this by fetching the state of the scheduler and by reading the logs that (per job) get written to `$HOME/.cheek/`. Note that you can ignore these logs, output of jobs will always go to stdout as well.

Note, `cheek` prior to version `0.3.0` originally used to boast a TUI, which has since been removed.

## Configuration

All configuration options are available by checking out `cheek --help` or the help of its subcommands (e.g. `cheek run --help`).

Configuration can be passed as flags to the `cheek` CLI directly. All configuration flags are also possible to set via environment variables. The following environment variables are available, they will override the default and/or set value of their similarly named CLI flags (without the prefix): `CHEEK_PORT`, `CHEEK_SUPPRESSLOGS`, `CHEEK_LOGLEVEL`, `CHEEK_PRETTY`, `CHEEK_HOMEDIR`.

## Events & Notifications

There are two types of event you can hook into: `on_success` and `on_error`. Both events materialize after an (attempted) job run. Three types of actions can be taken as a response: `notify_webhook`, `notify_slack_webhook` and `trigger_job`. See the example below. Definition of these event actions can be done on job level or at schedule level, in the latter case it will apply to all jobs.

```yaml
on_success:
  notify_webhook:
    - https://webhook.site/e33464a3-1a4f-4f1a-99d3-743364c6b10f
jobs:
  coffee:
    command: this fails # this will create on_error event
    cron: "* * * * *"
    on_error:
      notify_webhook:
        - https://webhook.site/e33464a3-1a4f-4f1a-99d3-743364c6b10f
  beans:
    command: echo grind # this will create on_success event
    cron: "* * * * *"
```

Webhooks are a generic way to push notifications to a plethora of tools. There is a generic way to do this via the `notify_webhook` option or a Slack-compatible one via `notify_slack_webhook`.

The `notify_webhook` sends a JSON payload to your webhook url with the following structure:

```json
{
	"status": 0,
	"log": "I'm a teapot, not a coffee machine!",
	"name": "TeapotTask",
	"triggered_at": "2023-04-01T12:00:00Z",
	"triggered_by": "CoffeeRequestButton",
	"triggered": ["CoffeeMachine"] // this job triggered another one
}
```

The `notify_slack_webhook` sends a JSON payload to your Slack webhook url with the following structure (which is Slack app compatible):

```json
{
	"text": "TeapotTask (exitcode 0):\nI'm a teapot, not a coffee machine!"
}
```

## Acknowledgements

`cheek-turner` is building on top of many great OSS assets. Noteable thanks goes to:

- [chota](https://jenil.github.io/chota/): for a pico sized css framework
- [gronx](https://github.com/adhocore/gronx): for allowing me not to worry about CRON strings.

<br/>
 
![GitHub Contributors](https://contrib.rocks/image?repo=datarootsio/cheek)
