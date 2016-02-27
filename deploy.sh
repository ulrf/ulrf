go build -o orgs
rsync -avzh orgs root@ent:/var/www/web/orgs/orgs
ssh root@ent "sudo restart orgs"