# Chatlytics
A simple service for tracking number of clicks on a chat button in a shopiy store

## Usage
```sh
chatlytics --conn="username:password@tcp(mysql-host-addr)/db-name"
```

## Working
for all api calls made to /chat?shop_id=<shop_id>&url_path=<url_path> are logged into a mysql table with schema
```sql
CREATE TABLE `chat_click_event` (
  `shop_id` varchar(250) DEFAULT NULL,
  `hour` int(11) DEFAULT NULL,
  `day` int(11) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `url_path` varchar(250) DEFAULT NULL,
  UNIQUE KEY `shop_id` (`shop_id`,`day`,`hour`,`url_path`)
)```

number clicks are therefore aggregated hourly
