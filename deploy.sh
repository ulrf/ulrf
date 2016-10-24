go build -o orgs
goupx orgs
rsync -avzhp --progress orgs root@ent:/var/www/web/orgs/orgs
ssh root@ent "sudo restart orgs"