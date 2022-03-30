terraform {
  required_providers {
    vultr = {
      source  = "vultr/vultr"
      version = "2.10.0"
    }
  }
}

resource "vultr_ssh_key" "ssh_key" {
  name    = "my-ssh-key"
  ssh_key = file("${var.key}.pub")
}

resource "vultr_instance" "my_instance" {
  count       = var.num_inst
  plan        = var.plan
  region      = var.region
  app_id      = var.app
  ssh_key_ids = [vultr_ssh_key.ssh_key.id]

  provisioner "remote-exec" {
    script = "scripts/deploy.sh"

    connection {
      host        = self.main_ip
      private_key = file("${var.key}")
    }
  }
}


