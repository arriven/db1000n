
variable "region" {
  default = "icn" # Seoul, South Korea
}

variable "plan" {
  default = "vc2-1c-1gb"
}

variable "app" {
  description = "Docker Ubuntu 20.04"
  default     = "37"
}

variable "key" {
  description = "Path to SSH key"
  type        = string
}

variable "num_inst" {
  type = number
}
