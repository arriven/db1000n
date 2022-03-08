data "digitalocean_tag" "stop-sites" {
  name = var.tags
}

data "digitalocean_ssh_key" "terraform" {
  name = var.digitalocean_ssh_key
}

resource "local_file" "user_credentials" {
  content = templatefile("${path.module}/script.tpl", {
    db1000n_version = var.db1000n_version
  })
  filename = "${path.module}/script.sh"
}

resource "digitalocean_droplet" "db1000n" {
  for_each      = toset(var.regions)
  name          = "${var.name}-${each.key}"
  size          = var.size
  region        = each.key
  ipv6          = var.ipv6
  backups       = var.backups
  monitoring    = var.monitoring
  droplet_agent = var.droplet_agent
  image         = var.image_name

  tags = [data.digitalocean_tag.stop-sites.id]

  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]

  connection {
    user        = "root"
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
    host        = self.ipv4_address
  }

  provisioner "file" {
    source      = "${path.module}/script.sh"
    destination = "/opt/script.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /opt/script.sh",
      "/opt/script.sh",
    ]
  }

  depends_on = [ resource.local_file.user_credentials ]
}
