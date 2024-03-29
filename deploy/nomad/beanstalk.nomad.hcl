job "beanstalk" {
  group "beanstalk" {
    network {
      port "beanstalk" {
        to = 11300
      }
    }

    service {
      name = "beanstalk"
      port = "beanstalk"
    }

    task "beanstalk" {
      driver = "docker"

      config {
        image = "confa/beanstalk:latest"
        ports = ["beanstalk"]
      }

      resources {
        cpu    = 100
        memory = 16
        memory_max = 256
      }
    }
  }
}
