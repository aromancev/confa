job "avp" {
  group "avp" {
    task "avp" {
      driver = "docker"

      config {
        image = "confa/avp:latest"
        # This job processes video. It can consume all CPU kicking out other jobs.
        cpu_hard_limit = true
      }

      template {
        data = <<EOH
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_PROCESS_TRACK = "{{ key "beanstalk/tubes/process-track" }}"
          BEANSTALK_TUBE_UPDATE_RECORDING_TRACK = "{{ key "beanstalk/tubes/update-recording-track" }}"
          {{range service "minio" }}
            STORAGE_HOST = "{{.Address}}:{{.Port}}"
          {{end}}
          STORAGE_ACCESS_KEY = "{{ key "storage/access-key" }}"
          STORAGE_SCHEME = "http"
          STORAGE_REGION = "stub"
          STORAGE_SECRET_KEY = "{{ key "storage/secret-key" }}"
          STORAGE_BUCKET_TRACK_RECORDS = "{{ key "storage/buckets/confa-tracks-internal" }}"
          STORAGE_BUCKET_TRACK_PUBLIC = "{{ key "storage/buckets/confa-tracks-public" }}"
        EOH
        destination = "secrets/.env"
        env         = true
      }

      resources {
        cpu    = 700
        memory = 256
        memory_max = 512
      }
    }
  }
}
