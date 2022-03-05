resource "google_service_account" "vm" {
  project      = var.project_id
  account_id   = "compute-sa"
  display_name = "Service Account for compute engine"
}

resource "google_project_iam_member" "vm_logs" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.vm.email}"
}

resource "google_project_iam_member" "vm_metric" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.vm.email}"
}

resource "google_compute_instance" "core" {
  count = 1

  project      = var.project_id
  name         = "atck-${count.index}"
  machine_type = var.machine_type
  zone         = var.machine_location

  tags = ["default-allow-ssh"]

  boot_disk {
    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/ubuntu-minimal-2004-focal-v20220203"
    }
  }

  network_interface {
    network = "default"
    access_config {}
  }

  scheduling {
    preemptible       = true
    automatic_restart = false
  }

  service_account {
    email  = google_service_account.vm.email
    scopes = ["cloud-platform", "logging-write", "monitoring-write"]
  }

  metadata_startup_script = <<EOT
apt-get remove docker docker-engine docker.io containerd runc
apt-get update
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    cron
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

ulimit -n 30000
ulimit -n 30000

wget -q "https://www.expressvpn.works/clients/linux/expressvpn_3.18.1.0-1_amd64.deb" -O /tmp/expressvpn_3.18.1.0-1_amd64.deb
dpkg -i /tmp/expressvpn_3.18.1.0-1_amd64.deb \
    && rm -rf /tmp/*.deb

cat <<EOF >> ./activate.sh
#!/usr/bin/expect
spawn expressvpn activate
expect "code:"
send "${var.expressvpn_key}\r"
expect "information."
send "n\r"
expect eof
EOF
chmod +x ./activate.sh && ./activate.sh

expressvpn preferences set send_diagnostics false
expressvpn preferences set auto_connect true
expressvpn connect "${var.vpn_location}"

sleep 10
echo "IP HERE-> $(curl -v ifconfig.me)"

cat <<EOF >> ./run.sh
#! /bin/bash
docker stop $(docker ps -a -q)
docker run --rm ghcr.io/arriven/db1000n:latest
EOF
chmod +x ./run.sh

(crontab -l ; echo '*/10 * * * * /usr/bin/sudo /run.sh >> /var/log/run.log 2>&1') | crontab -

EOT
}
