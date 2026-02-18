# QUEUE PROCESSING
- complete queue functionality so that i can call the queueAdapter.enqueue() for any of the queue with the necessary params from anywhere in my code. an example in JS is 
```
await this.jobQueueAdapter.addToQueue(QUEUES.INBOX_QUEUE, {
    data: {...},
    event: InboxEventTypeEnum.BULK_PRODUCT_UPLOAD,
});
```
- i want to have two queues which is the outbox and inbox queue, diffrerent kind of jobs can go into any of the queue but in the consumer, there would be a factory, that based on the type of event that was enqueue, the factory would know the usecase to call and execute. 
- then inside the choosen usecase, it would gather it's own dependencies and start processing the job.
  
# SENDING EMAILS
- The same way i have my storage setup, where i can just call the storage adapter to upload the image and the storage adapter  would know if it is to use s3 or azure blob storage or whatever provider. I want to have an email adapter that would call the email provider that is is configured to use. which can be sendgrid, smtp, outlook, etc

# CRON
- i want to be able to implement a cron for calculating different tasks accordin to Week 23 in @month6.md

- i also want you to implement webhook, pdf export, csv export and the concurrency features in month 7.md 

Do you have any question?

