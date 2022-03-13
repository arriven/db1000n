variable "project_id" {
  description = "GCP project id"
}

variable "expressvpn_key" {
  description = "expressvpn activation key"
}

variable "machine_type" {
  description = "GCP machine type"
  default     = "n1-standard-1"
}

variable "machine_location" {
  description = "machine location"
  default     = "us-central1-a"
}

variable "machine_count" {
  description = "how many VM's will be created"
  default     = 2
}