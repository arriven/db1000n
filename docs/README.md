# Death by 1000 needles

Developed by Bohdan Ivashko (https://github.com/Arriven)

## Інструкція для новачків як встановити на виділений сервер google clouds

Спочатку треба зареєструватися та створити виділений сервер в cloud.google.com або іншому сервісі.
cloud.google.com дають при старті 300$ для використання серверів - цього повинно хватати надовго.
Робимо все по інструкції як вказано тут:

 **https://telegra.ph/%D0%86nstrukc%D1%96ya-yak-DDositi-sajti-za-dopomogoyu-server%D1%96v-02-26** 

тільки в пункті 6 вибираємо **e2-medium** тому що **e2-micro** не тягне. 

доходимо до пункту 8.
далі робимо слідуюче:

### Спочатку оновлюємо систему

для копіювання і вставлення команд в командному рядку працюють комбінації клавіш **ctrl**+**C** **ctrl**+**V**

```
sudo apt-get update
```
потім встановлюємо программи (виконуємо команди по одній)
```
sudo apt-get install screen
```
```
sudo apt-get install openvpn
```
попросить підтвердження встановлення. нажимаємо **y** та **enter**



### Создаємо конфіг для впн
```
nano autovpn3.sh
```         

відкриється редактор, копіюємо і вставляємо туди текст нижче. **ctrl**+**C** **ctrl**+**V**

```
#!/bin/bash
 
# autovpn3, coded by MiAl, 
# you can leave a bug report on the page: https://miloserdov.org/?p=5858
# сообщить об ошибке на русском вы можете на странице: https://HackWare.ru/?p=15429
 
# you can change these parameters:
country='RU' # empty for any or JP, KR, US, TH, etc.
useSavedVPNlist=0 # set to 1 if you don't want to download VPN list every time you restart this script, otherwise set to 0
useFirstServer=0 # set the value to 0 to choose a random VPN server, otherwise set to 1 (maybe the first one has higher score)
vpnList='/tmp/vpns.tmp'
proxy=0 # replace with 1 if you want to connect to VPN server through a proxy
proxyIP=''
proxyPort=8080
proxyType='socks' # socks or http
 
# don't change this:
counter=0
VPNproxyString=''
cURLproxyString=''
 
if [ $proxy -eq 1 ];then
    echo 'We will use a proxy'
    if [ -z "$proxyIP" ]; then
        echo "To use a proxy, you must specify the proxy's IP address and port (hardcoded in the source code)."
        exit
    else
        if [ "$proxyType" == "socks" ];then
            VPNproxyString=" --socks-proxy $proxyIP $proxyPort "
            cURLproxyString=" --proxy socks5h://$proxyIP:$proxyPort "
        elif [ "$proxyType" == "http" ];then
            VPNproxyString=" --http-proxy $proxyIP $proxyPort "
            cURLproxyString=" --proxy http://$proxyIP:$proxyPort "
        else
            echo 'Unsupported proxy type.'
            exit   
        fi 
    fi
fi
 
if [ $useSavedVPNlist -eq 0 ];then
    echo 'Getting the VPN list'
    curl -s $cURLproxyString https://www.vpngate.net/api/iphone/ > $vpnList
elif [ ! -s $vpnList ];then
    echo 'Getting the VPN list'
    curl -s $cURLproxyString https://www.vpngate.net/api/iphone/ > $vpnList  
else
    echo 'Using existing VPN list'
fi
 
while read -r line ; do
    array[$counter]="$line"
    counter=$counter+1
done < <(grep -E ",$country" $vpnList)
 
CreateVPNConfig () {
    if [ -z "${array[0]}" ]; then
        echo 'No VPN servers found from the selected country.'
        exit
    fi
 
    size=${#array[@]}
 
    if [ $useFirstServer -eq 1 ]; then
        index=0
        echo ${array[$index]} | awk -F "," '{ print $15 }' | base64 -d > /tmp/openvpn3
    else       
        index=$(($RANDOM % $size))
        echo ${array[$index]} | awk -F "," '{ print $15 }' | base64 -d > /tmp/openvpn3
    fi
 
    echo 'Choosing a VPN server:'
    echo "Found VPN servers: $((size+1))"
    echo "Selected: $index"
    echo "Country: `echo ${array[$index]} | awk -F "," '{ print $6 }'`"   
}
 
while true
    do
        CreateVPNConfig
        echo 'Trying to start OpenVPN client'
        sudo openvpn --config /tmp/openvpn3 $VPNproxyString
        read -p "Try another VPN server? (Y/N): " confirm && [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]] || exit
    done
```	
	
	
для вихода нажимаємо **ctrl**+**X** далі для збереження нажимаємо **Y** та **enter**

даємо права на виконання файлу
```
chmod +x autovpn3.sh
```

### Далі запускаємо впн.

создаємо нову сессію терміналу

```
screen
```             

видасть додатокве вікно натискаємо там **enter**


запускаємо скрипт vpn

```
sudo ./autovpn3.sh
```

якщо видає помилку ***No VPN servers found from the selected country.***
то відкриваємо ще раз конфіг на редагування 

```
nano autovpn3.sh
```

ідемо на строку ***country='RU' ***
і змінюємо країну на іншу. наприклад японія **JP**


запуститься скрипт, щоб вийти з цієї сессії терміналу потрібно натиснути **ctrl+A ctrl+D** 

### Далі встановлюємо докер

```
 wget https://gist.githubusercontent.com/antl31/83229c0eaaa1d259a569cfb57ab75230/raw/b4e01a106fbe534a7bfa4930a66cf933c6366c5c/install_docker.sh
```

```
bash install_docker.sh
```

встановиться докер та відкриється вже нова сесія терміналу. натискаємо **enter**


### Запускаємо скрипт

```
sudo docker run ghcr.io/arriven/db1000n
```

якщо скрипт запустився та працює постійно хочаб хвилин 10 то все нормально. профіт. 
якщо скрипт попрацював трохи і вийшов - змінюйте країну і запускайте наново.



щоб вийти з цієї сессії терміналу потрібно натиснути **ctrl+A ctrl+D** 


коли ви вийдете то можете закривати вікно все працює. 
через деякий час щоб перевірити як працює скрипт треба ще раз відкрити через пункт **Open in browser window** в панелі керування гугл клоуд 

тут треба навчитися користуватися вікнами **screen**

пишемо 
```
screen -ls
```


побачимо вивід команди:

**There are screens on:**
      **4688.pts-0.instance-5   (03/04/22 11:12:13)     (Detached)**
      **2219.pts-0.instance-5   (03/04/22 09:59:36)     (Detached)**
**2 Sockets in /run/screen**



тут ми бачимо що запущені 2 сесії. номери в вас будуть інші. дивимося по часу запуску що одна запущена раніше, інша пізніше.
значить та шо раніше там впн, та шо пізніше - наша программа для ддос.
щоб перейти на вікно з программою потрібно написати
screen -r 4688

номер у вас буде свій. вписуєте той шо в вас. 
ви перейдете на термінал з запущеною программою.



якщо у вас виникають проблеми із запуском то зверніться в чат телеграмм https://t.me/+md-77jub3zQyOWRi





```bash
go install github.com/Arriven/db1000n@latest
~/go/bin/db1000n
```

### docker install

how to install docker?

https://docs.docker.com/get-docker/

make sure you've set all available resources to docker

https://docs.docker.com/desktop/windows/#resources
https://docs.docker.com/desktop/mac/#resources

run d1000n

```bash
docker run --pull always ghcr.io/arriven/db1000n:latest
```

### kubernetes install

for experienced users, see instructions in [helm/](helm/)

### shell install

run install script directly into the shell (useful for install through ssh)

```bash
curl https://raw.githubusercontent.com/Arriven/db1000n/main/install.sh | sh
```

the command above will detect the os and architecture, dowload the archive, validate it, and extract db1000n executable into the working directory. You can then run it via this command

```bash
./db1000n
```

### k8s run

Enter k8s directory

```bash
$ cd k8s-manifest
```

Create deployment

```bash
$ kubectl create -f ./
```

Scale it, if you have resources

```bash
$ kubectl scale deployment/db1000n --replicas=10 -n db1000n
```

## Configuration

### Commandline reference

```text
Usage of /tmp/go-build781992389/b001/exe/main:
  -b string
        path to a backup config file in case primary one is unavailable (default "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json")
  -c string
        path to a config file, can be web endpoint (default "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json")
  -h    print help message and exit
  -l int
        logging level. 0 - Debug, 1 - Info, 2 - Warning, 3 - Error (default 1)
  -m string
        path where to dump usage metrics, can be URL or file, empty to disable
  -r duration
        refresh timeout for updating the config (default 1m0s)
```

### Config file reference

The config is expected to be in json format and has following configuration values:

- `jobs` - `[array]` array of attack job definitions to run, should be defined inside the root object
- `jobs[*]` - `[object]` single job definition as json object
- `jobs[*].type` - `[string]` type of the job (determines whhich attack function to launch). Can be `http`, `tcp`, `udp`, `syn-flood`, or `packetgen`
- `jobs[*].count` - `[number]` the amount of instances of the job to be launched, automatically set to 1 if no or invalid value is specified
- `jobs[*].args` - `[object]` arguments to pass to the job. Depends on `jobs[*].type`

`http` args:

- `method` - `[string]` http method to use (passed directly to go `http.NewRequest`)
- `path` - `[string]` url path to use (passed directly to go `http.NewRequest`)
- `body` - `[object]` http payload to use (passed directly to go `http.NewRequest`)
- `headers` - `[object]` key-value map of http headers
- `client` - `[object]` http client config for the job
- `client.tls_config` - `[object]` tls config for transport (InsecureSkipVerify is true by default)
- `client.proxy_urls` - `[array]` string urls for proxies to use (chosen randomly for each request)
- `client.timeout` - `[time.Duration]`
- `client.max_idle_connections` - `[number]`

`tcp` and `udp` shared args:

- `address` - `[string]` network host to connect to, can be either `hostname:port` or `ip:port`
- `body` - `[object]` json data to be repeatedly sent over the network

`http`, `tcp`, and `udp` shared args:

- `interval_ms` - `[number]` interval between requests in milliseconds. Defaults to 0 (Care, in case of udp job it might generate the data faster than your OS/network card can process it)
- `count` - `[number]` limit the amount of requests to send with this job invocation. Defaults to 0 (no limit). Note: if config is refreshed before this limit is reached the job will be restarted and the counter will be reset

`syn-flood` args:

- `host` - `[string]` host to attack, can be either DNS name or IP
- `port` - `[number]` port to attack
- `payload_length` - `[number]` refer to original syn-flood package docs
- `flood_type` - `[string]` type of flood to send, can be `syn`, `ack`, `synack`, and `random`

Warning: `packetgen` requires root privilleges to run

`packetgen` args:

- `host` - `[string]` host to attack, can be either DNS name or IP
- `port` - `[string]` numerical value of port to attack (string to allow template generation)
- `payload` - `[string]` payload to include into packets
- `ethernet` - `[object]` ethernet layer configuration
- `ethernet.src_mac` - `[string]`
- `ethernet.dst_mac` - `[string]`
- `ip` - `[object]` ip layer configuration
- `ip.src_ip` - `[string]`
- `ip.dst_ip` - `[string]`
- `udp` - `[object]` udp layer configuration (disables tcp if present)
- `udp.src_port` - `[number]`
- `udp.dst_port` - `[number]`
- `tcp` - `[object]` tcp layer configuration (excluded if udp is present)
- `tcp.src_port` - `[number]`
- `tcp.dst_port` - `[number]`
- `tcp.seq` - `[number]`
- `tcp.ack` - `[number]`
- `tcp.window` - `[number]`
- `tcp.urgent` - `[number]`
- `tcp.flags` - `[object]` flags for tcp (every flag has it's respective name)

Warning: `slow-loris` from testconfig.json is not yet finished and may overload the app due to not handling config refreshes

Almost every leaf `[string]` or `[object]` parameter can be templated with go template syntax. I've also added couple helper functions (list will be growing):

- `random_uuid`
- `random_int`
- `random_int_n`
- `random_ip`
- `random_payload`
- `random_mac_addr`
- `random_port`
- `local_ip`
- `local_mac_addr`
- `base64_encode`
- `base64_decode`

Please refer to official go documentation and code for these for now
