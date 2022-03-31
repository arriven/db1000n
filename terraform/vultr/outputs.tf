output "ip" {
  value = vultr_instance.my_instance[*].main_ip
}
