resource "hcloud_ssh_key" "key" {
  name       = "db1000n_key"
  public_key = var.ssh_public_key
}

resource "hcloud_server" "server" {
  count       = var.instance_count
  name        = "db1000n-server-${count.index}"
  image       = var.os_type
  server_type = var.server_type
  location    = var.location
  ssh_keys    = [hcloud_ssh_key.key.id]

  labels = {
    app = "db1000n"
  }

  user_data = file("user_data.yml")
}
