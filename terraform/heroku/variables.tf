variable "email" {
  type = string
}

variable "api_key" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "eu"
}

variable "repo" {
  type    = string
  default = "https://github.com/Arriven/db1000n"
}

variable "app_version" {
  type = string
}

variable "instance_count" {
  type    = number
  default = 1
}

# https://devcenter.heroku.com/articles/dyno-types
variable "instance_type" {
  type    = string
  default = "free"
}

variable "config_path" {
  type    = string
  default = "https://gist.githubusercontent.com/ddosukraine2022/f739250dba308a7a2215617b17114be9/raw/db1000n_targets.json"
}
