#!/bin/bash
KEYNAME="$1"
echo "-----> Booting EC2 instance..."
INSTANCE=$(ec2-run-instances -k $1 ami-c90277a0 2>/dev/null | awk '/^INSTANCE/ {print $2}')
terminate() {
  echo "-----> Terminating $INSTANCE..."
  ec2-terminate-instances $INSTANCE &>/dev/null && echo "       Shutting down"
}
#trap "terminate" EXIT
sleep 30
status=""
while [[ "$status" != "running" ]]; do
    info=$(ec2-describe-instances 2>/dev/null | grep $INSTANCE)
    status=$(echo "$info" | cut -f 6 | grep run)
    echo "       Waiting..."
    sleep 5
    if [[ $status == "running" ]]; then
        echo "-----> $INSTANCE has succesfully booted!"
        break
    fi
done
PUBLIC_IP=$(echo "$info" | awk '{print $14}')
echo "-----> Waiting for SSH on instance..."
sleep 10
echo "-----> Connecting and running boostrap script..."
indent() { sed "s/^/       /"; }
cat ../bootstrap.sh | ssh -o "StrictHostKeyChecking=no" ubuntu@$PUBLIC_IP "HOSTNAME=$PUBLIC_IP sudo bash" 2>/dev/null | indent