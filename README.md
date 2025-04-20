# Belabox Cloud stats exporter

Simple Belabox Cloud prometheus exporter

### Just run
```shell
go run belabox-exporter.go
```

### Build
```shell
go build -o exporter belabox-exporter.go
```

### Build for docker
```shell
docker build . -t blackvanilla/belacloud:latest
```

### Or simply pull image
```shell
docker pull blackvanilla/belacloud:latest
docker run --rm -p 9090:9090 blackVanilla/belacloud:latest
```

### Get prometheus metrics
```shell
curl http://localhost:9090/probe?url=http://foo.srt.belabox.net:8080/cH9aN7gE0T1hI5s8K3eY7uWqLpOd
```

### Configure prometheus scrape config
```yaml
- job_name: 'belacloud_exporter'
  metrics_path: /probe

  static_configs:
    - targets:
        - belacloud:9090
      labels:
        __param_url: http://foo.srt.belabox.net:8080/cH9aN7gE0T1hI5s8K3eY7uWqLpOd
        __param_name: "murr1to foo"
        instance: belabox_cloud_foo

  relabel_configs:
    - source_labels: [__param_url]
      target_label: url
    - source_labels: [__param_name]
      target_label: name
    - source_labels: [__address__]
      target_label: instance
```