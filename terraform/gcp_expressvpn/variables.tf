variable "project_id" {
  description = "GCP project id"
}

variable "expressvpn_key" {
  description = "expressvpn activation key"
}

variable "vpn_location" {
  description = "vpn server location to use initially"
}

variable "machine_type" {
  description = "GCP machine type"
  default     = "n1-standard-1"
}

variable "machine_location" {
  description = "machine location"
  default     = "us-central1-a"
}
