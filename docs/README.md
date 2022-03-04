# Death by 1000 needles

Developed by Arriven (https://github.com/Arriven)

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

якщо видає помилку **_No VPN servers found from the selected country._**
то відкриваємо ще раз конфіг на редагування

```
nano autovpn3.sh
```

ідемо на строку **_country='RU' _**
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

**There are screens on:**<br/>
**4688.pts-0.instance-5 (03/04/22 11:12:13) (Detached)**<br/>
**2219.pts-0.instance-5 (03/04/22 09:59:36) (Detached)**<br/>
**2 Sockets in /run/screen**

тут ми бачимо що запущені 2 сесії (номери на початку 2 і 3 строки у вас будуть інші).
дивимося по часу запуску що одна запущена раніше, інша пізніше.
значить та шо раніше там впн, та шо пізніше - наша программа для ддос.
щоб перейти на вікно з программою потрібно написати

```
screen -r 4688
```

номер у вас буде свій. вписуєте той шо в вас.
ви перейдете на термінал з запущеною программою.

якщо у вас виникають проблеми із запуском то зверніться в чат телеграмм https://t.me/+md-77jub3zQyOWRi
