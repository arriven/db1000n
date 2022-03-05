rem First we write PowerShell script to separate file, since it's easier to do download from github with powershell
rem Parsing command line argument to get target download folder
echo $SaveFolder = $args[0]                                                       >  %temp%\GetDB1000N.ps1
echo if ($args.count -lt 1)                                                       >> %temp%\GetDB1000N.ps1
echo {                                                                            >> %temp%\GetDB1000N.ps1
echo 	Write-Host "Missing save folder parameter"                                >> %temp%\GetDB1000N.ps1
echo }                                                                            >> %temp%\GetDB1000N.ps1
rem Create target folder if it don't exist yet
echo New-Item -ItemType Directory -Force -Path $SaveFolder                        >> %temp%\GetDB1000N.ps1
rem Getting list of files in latest release via github API and parse response JSON
echo $JSONURL = "https://api.github.com/repos/Arriven/db1000n/releases/latest"    >> %temp%\GetDB1000N.ps1
echo $JSON = Invoke-WebRequest -Uri $JSONURL                                      >> %temp%\GetDB1000N.ps1
echo $ParsedJSON = ConvertFrom-Json -InputObject $JSON                            >> %temp%\GetDB1000N.ps1
echo $Assets = Select-Object -InputObject $ParsedJSON -ExpandProperty assets      >> %temp%\GetDB1000N.ps1
rem Iterate over list of all files in release
echo Foreach ($Asset IN $Assets)                                                  >> %temp%\GetDB1000N.ps1
echo {                                                                            >> %temp%\GetDB1000N.ps1
rem Search for windows x64 build with regex
echo 	if ($Asset.name -match 'db1000n-.*-windows-amd64.zip')                    >> %temp%\GetDB1000N.ps1
echo 	{                                                                         >> %temp%\GetDB1000N.ps1
rem Download found build
echo 		$DownloadURL = $Asset.browser_download_url                            >> %temp%\GetDB1000N.ps1
echo 		$ZIPPath = Join-Path -Path $SaveFolder -ChildPath "db1000n.zip"       >> %temp%\GetDB1000N.ps1
echo 		Invoke-WebRequest -Uri $DownloadURL -OutFile $ZIPPath                 >> %temp%\GetDB1000N.ps1
rem Extract downloaded archive
echo 		Expand-Archive -Path $ZIPPath -DestinationPath $SaveFolder -Force     >> %temp%\GetDB1000N.ps1
rem Delete original archive file
echo 		Remove-Item -Path $ZIPPath                                            >> %temp%\GetDB1000N.ps1
echo 		Write-Host "Sucessfully downloaded DB1000N $($Asset.name)"            >> %temp%\GetDB1000N.ps1
echo 		exit 0                                                                >> %temp%\GetDB1000N.ps1
echo 	}                                                                         >> %temp%\GetDB1000N.ps1
echo }                                                                            >> %temp%\GetDB1000N.ps1
echo Write-Host "Something went wrong, couldn't download DB1000N"                 >> %temp%\GetDB1000N.ps1
echo exit 1                                                                       >> %temp%\GetDB1000N.ps1

rem Download latest windows x64 build
rem We don't currently check if it's already downloaded, so it redownload latest version each run
powershell -ExecutionPolicy Bypass -File %temp%\GetDB1000N.ps1 %temp%
rem Assume that archive contained executable file named db1000n.exe and try to run it
:StartApp
rem Starts the app and if it returns 0 exit code it does "goto End", breaking loop
%temp%/db1000n.exe && goto End
rem Otherwise, if exit code is non zero, report and restart
echo DB1000N crahed with exit code %errorlevel%, restarting
goto StartApp
:End
exit /b 0
