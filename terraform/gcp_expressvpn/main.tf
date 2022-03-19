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

resource "google_compute_instance_template" "atck" {
  name         = "atck-template"
  machine_type = var.machine_type
  tags         = ["default-allow-ssh"]

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

cat <<EOF >> ./countries.txt
Hong Kong
Singapore
India
Canada
Japan
Germany
Mexico
Australia
United Kingdom
Netherlands
Spain
South Korea
Switzerland
France
Philippines
Malaysia
Sri Lanka
Italy
Pakistan
Kazakhstan
Thailand
Indonesia
Taiwan
Vietnam
Macau
Cambodia
Mongolia
Laos
Myanmar
Nepal
Kyrgyzstan
Uzbekistan
Bangladesh
Bhutan
Brazil
Panama
Chile
Argentina
Bolivia
Colombia
Venezuela
Ecuador
Guatemala
Peru
Uruguay
Bahamas
Sweden
Romania
Turkey
Ireland
Iceland
Norway
Denmark
Belgium
Greece
Portugal
Austria
Finland
EOF

sudo docker run \
--env=ACTIVATION_CODE=${var.expressvpn_key} \
--env=PREFERRED_PROTOCOL=auto \
--env=LIGHTWAY_CIPHER=auto \
--env=SERVER=$(shuf -n 1 /countries.txt) \
-e NETWORK=192.168.1.0/24 \
--cap-add=NET_ADMIN \
--device=/dev/net/tun \
--privileged \
--tty=true \
--name=vpn \
--detach=true \
--dns=1.1.1.1 \
--tty=true \
polkaned/expressvpn \
/bin/bash

sleep 10

cat <<EOF >> ./run.sh
#! /bin/bash
docker stop db1000n
docker exec vpn expressvpn disconnect
docker exec vpn expressvpn connect "$(shuf -n 1 /countries.txt)"
docker run --name=db1000n --pull=always --net=container:vpn -e PUID=1000 -e PGID=1000 --log-driver=gcplogs --rm -d ghcr.io/arriven/db1000n-advanced:latest
EOF
chmod +x ./run.sh

(crontab -l ; echo '*/10 * * * * /usr/bin/sudo /run.sh') | crontab -

docker run --name=db1000n --net=container:vpn -e PUID=1000 -e PGID=1000 --log-driver=gcplogs --rm -d ghcr.io/arriven/db1000n-advanced:latest

EOT

  service_account {
    email  = google_service_account.vm.email
    scopes = ["cloud-platform", "logging-write", "monitoring-write"]
  }

  network_interface {
    network = "default"
    access_config {}
  }

  disk {
    source_image = "projects/ubuntu-os-cloud/global/images/ubuntu-minimal-2004-focal-v20220203"
    auto_delete  = true
    boot         = true
  }

  scheduling {
    preemptible       = true
    automatic_restart = false
  }
}

resource "google_compute_instance_group_manager" "attckrs" {
  name               = "attckrs"
  base_instance_name = "atck"
  zone               = var.machine_location
  target_size        = var.machine_count

  version {
    instance_template = google_compute_instance_template.atck.id
  }
}
