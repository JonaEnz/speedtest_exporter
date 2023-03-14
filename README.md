# speedtest_exporter

![License](https://img.shields.io/github/license/JonaEnz/speedtest_exporter)

This Prometheus exporter check your network connection. Metrics are :

* Latency
* Download bandwidth
* Upload bandwidth


## Installation

Build and run the binary using go >= 1.18.

## Usage

Launch the Prometheus exporter :

```bash
$ speedtest_exporter -log.level=debug
```

## Local Deployment

* Launch Prometheus using the configuration file in this repository:

```bash
$ prometheus -config.file=prometheus.yml
```

* Launch exporter:

```bash
$ speedtest_exporter -log.level=debug
```

* Check that Prometheus find the exporter on `http://localhost:9090/targets`

## License

See [LICENSE](LICENSE) for the complete license.


## Changelog

A [changelog](ChangeLog.md) is available
