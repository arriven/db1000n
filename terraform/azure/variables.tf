variable "bomblet_count" {
  type        = number
  default     = 1
  description = "Number of containers per region."
}

variable "prefix" {
  default     = "attack"
  description = "The default prefix for resources."
}