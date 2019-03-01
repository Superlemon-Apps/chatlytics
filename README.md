# Chatlytics
A simple service for tracking number of clicks on a chat button in a shopify store

## installation
```sh
go get github.com/sankalpjonn/chatlytics
```

## Usage
```sh
chatlytics --conn="username:password@tcp(mysql-host-addr)/db-name"
```

## Working
for all api calls made to /chat?shop_id=<shop_id>&url_path=<url_path> are logged into a mysql table with schema

click counts are aggragated hourly for a particular shop, day and url_path

```sql
CREATE TABLE chat_click_event (
  shop_id varchar(250) DEFAULT NULL,
  hour int(11) DEFAULT NULL,
  day int(11) DEFAULT NULL,
  count int(11) DEFAULT NULL,
  url_path varchar(250) DEFAULT NULL,
  UNIQUE KEY shop_id (shop_id, day, hour , url_path)
)```
