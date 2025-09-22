#!/usr/bin/bash

# this script does not work for Basic SKU Service Bus
# Supported Service Bus SKUs is Standard and Premium

resource_group=""
namespace=""
main_queue=""
fwdr_queue="${main_queue}-fwdr"
 
# create fwdr queue
echo "creating forwarder queue: $fwdr_queue"
az servicebus queue create \
  --resource-group $resource_group \
  --namespace-name $namespace \
  --name "$fwdr_queue" \
  --forward-to $main_queue
 
sleep 1
 
# enable main queue's DLQ forwarder
echo "enabling DLQ message forwarder on $main_queue to $fwdr_queue"
az servicebus queue update \
  --resource-group $resource_group \
  --namespace-name $namespace \
  --name $main_queue \
  --forward-dead-lettered-messages-to "$fwdr_queue"
 
sleep 1
 
# wait for main queue's DLQ messages to be forwarded to fwdr queue
while true; do
    echo "checking for DLQ messages in $main_queue"
    dlq_message_count=$(az servicebus queue show \
        --resource-group $resource_group \
        --namespace-name $namespace \
        --name $main_queue \
        --query "countDetails.deadLetterMessageCount" -o tsv)
 
    if [ "$dlq_message_count" -gt 0 ]; then
        echo "$main_queue has $dlq_message_count DLQ messages left, checking again in 3 seconds..."
        sleep 3
    else
        echo "$main_queue has no DLQ messages left"
        break
    fi
done
 
echo "sleeping 10 seconds..."
sleep 10
 
# disable main queue's DLQ forwarder
echo "disabling DLQ message forwarder on $main_queue"
az servicebus queue update \
    --resource-group $resource_group \
    --namespace-name $namespace \
    --name $main_queue \
    --forward-dead-lettered-messages-to ""
 
# wait for fwdr queue's messages to be forwarded to main queue
while true; do
    echo "checking for messages in $fwdr_queue"
    fwdr_message_count=$(az servicebus queue show \
        --resource-group $resource_group \
        --namespace-name $namespace \
        --name "$fwdr_queue" \
        --query "countDetails.activeMessageCount" -o tsv)
 
    if [ "$fwdr_message_count" -gt 0 ]; then
        echo "$fwdr_queue has $fwdr_message_count messages left, checking again in 3 seconds..."
        sleep 3
    else
        echo "$fwdr_queue has no messages left"
        break
    fi
done
 
echo "sleeping 10 seconds..."
sleep 10
 
# delete fwdr queue
echo "deleting forwarder queue: $fwdr_queue"
az servicebus queue delete \
    --resource-group $resource_group \
    --namespace-name $namespace \
    --name "$fwdr_queue"
 
echo "done"