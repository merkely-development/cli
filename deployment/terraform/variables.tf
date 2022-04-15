variable "reporter_apps" {
  type = map(any)
  default = {
    staging = {
      merkely_host = "https://staging.app.merkely.com"
      cpu_limit = 100
      mem_limit = 450
      mem_reservation = 64
    }
    prod = {
      merkely_host = "https://app.merkely.com"
      cpu_limit = 100
      mem_limit = 450
      mem_reservation = 64
    }
  }
}

variable "env" {
  type = string
}

variable "merkely_env" {
  type = string
}

variable "app_name" {
  type    = string
  default = "merkely-cli"
}

variable "ecr_replication_targets" {
  type    = list(map(string))
  default = []
}

variable "ecr_replication_origin" {
  type    = string
  default = ""
}

variable "ecs_sluster_name" {
  type    = string
  default = "merkely-reporter"
}

variable "TAGGED_IMAGE" {
  type = string
}
