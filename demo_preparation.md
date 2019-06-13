# bluebutton POETic DEMO

## Terminal Session 1:
``` bash
cd ~/Documents/_Local/Developer/poet_dcr/bluebutton-web-server/
docker-compose up web

```

## Terminal Session 2:

``` bash
cd ~/Documents/_Local/Developer/go/src/github.com/whytheplatypus/bbdcr

# this command replaces the certification_Callback with the address of localhost that is running the go app.
./get_container_host_ip

rm certification.txt

go run main.go


```

## Open Safari Browser
http://localhost:8000/v1/certification/requests/

Get the last entry ID

[http://localhost:8000/v1/certification/certify/{ID}/](http://localhost:8000/v1/certification/certify/10/)

## Open Chrome Browser

http://localhost:8080

- Allow APP
- Display Patient record
- 


# Prep for Demo
[http://localhost:8000/admin/certification/certificationrequest/](http://localhost:8000/admin/certification/certificationrequest/)

Delete certification requests


[http://localhost:8000/admin/dcr/softwarestatement/](http://localhost:8000/admin/dcr/softwarestatement/)
Delete Softwarestatement
