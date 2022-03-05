echo $SaveFolder = $args[0]                                                       >  %temp%\GetDB1000N.ps1
echo if ($args.count -lt 1)                                                       >> %temp%\GetDB1000N.ps1
echo {                                                                            >> %temp%\GetDB1000N.ps1
echo 	Write-Host "Missing save folder parameter"                                >> %temp%\GetDB1000N.ps1
echo }                                                                            >> %temp%\GetDB1000N.ps1
echo New-Item -ItemType Directory -Force -Path $SaveFolder                        >> %temp%\GetDB1000N.ps1
echo $JSONURL = "https://api.github.com/repos/Arriven/db1000n/releases/latest"    >> %temp%\GetDB1000N.ps1
echo $JSON = Invoke-WebRequest -Uri $JSONURL                                      >> %temp%\GetDB1000N.ps1
echo $ParsedJSON = ConvertFrom-Json -InputObject $JSON                            >> %temp%\GetDB1000N.ps1
echo $Assets = Select-Object -InputObject $ParsedJSON -ExpandProperty assets      >> %temp%\GetDB1000N.ps1
echo Foreach ($Asset IN $Assets)                                                  >> %temp%\GetDB1000N.ps1
echo {                                                                            >> %temp%\GetDB1000N.ps1
echo 	if ($Asset.name -match 'db1000n-.*-windows-amd64.zip')                    >> %temp%\GetDB1000N.ps1
echo 	{                                                                         >> %temp%\GetDB1000N.ps1
echo 		$DownloadURL = $Asset.browser_download_url                            >> %temp%\GetDB1000N.ps1
echo 		$ZIPPath = Join-Path -Path $SaveFolder -ChildPath "db1000n.zip"       >> %temp%\GetDB1000N.ps1
echo 		Invoke-WebRequest -Uri $DownloadURL -OutFile $ZIPPath                 >> %temp%\GetDB1000N.ps1
echo 		Expand-Archive -Path $ZIPPath -DestinationPath $SaveFolder -Force     >> %temp%\GetDB1000N.ps1
echo 		Remove-Item -Path $ZIPPath                                            >> %temp%\GetDB1000N.ps1
echo 		Write-Host "Sucessfully downloaded DB1000N $($Asset.name)"            >> %temp%\GetDB1000N.ps1
echo 		exit 0                                                                >> %temp%\GetDB1000N.ps1
echo 	}                                                                         >> %temp%\GetDB1000N.ps1
echo }                                                                            >> %temp%\GetDB1000N.ps1
echo Write-Host "Something went wrong, couldn't download DB1000N"                 >> %temp%\GetDB1000N.ps1
echo exit 1                                                                       >> %temp%\GetDB1000N.ps1

powershell -ExecutionPolicy Bypass -File %temp%\GetDB1000N.ps1 %temp%
%temp%/db1000n.exe