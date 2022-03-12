variable "bomblet_count" {
  type        = number
  default     = 1
  description = "Number of containers per region."
}

variable "prefix" {
  default     = "attack"
  description = "The default prefix for resources."
}

variable "attack_commands" {
  default     = ["/usr/src/app/db1000n","-c=https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json"]
  description = "The command to execute an attack with support of specifying additional flags."
}
