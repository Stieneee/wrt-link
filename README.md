# wrt-link

## Sample Set
```
cd /tmp/
echo 192.168.0.101 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC9vLRw/Gm75FLq+ekh6OaveP3v/3Bu+IYYP2rmmNWKTmjWqFZkNZGVlIzkhNZVM4v/st4NJFsEBJZLHUWcgtR/YKtA3pLW4kSZ7w3IPcu5kPVFg/swtDkz0rthLy61lNYWJmxw6azc9xZnSEow2/RdbcxMJLYnyNU6FQDIGSb7PGi34LW067FJlQ7ZjWazRxLdwTGtXvq39lmfolBg6tapDxOwu1XAewspxTBb0qQ9dc6Jkm8V7XOuom096qsUHguGTAXc/YaCvdS6/9zajj7TMQ4AFTcpsvsxu0vohFKfuCpSoH+9PPDWOs6FSLkis34amo3TnJckPqzv0YTZxOeL > /tmp/root/.ssh/known_hosts 
cat > /tmp/wrt-link.id_rsa <<- EOM
{{PASTE KEY}}
EOM
./wrt-link.sh test 192.168.0.101 2222
```
