variable "bomblet_count" {
  type        = number
  default     = 1
  description = "Number of containers per region."
}

variable "prefix" {
  default     = "main"
  description = "The default prefix for resources."
}

variable "attack_commands" {
  default     = null
  description = "The command to execute an attack with support of specifying additional flags."
}

variable "attack_environment_variables" {
  default     = { "ENABLE_PRIMITIVE" : "false" }
  description = "Environment variables for the container."
}
