CREATE DATABASE mocksvr;

CREATE TABLE `stub_interface` (
                                  `id` int(32) NOT NULL AUTO_INCREMENT,
                                  `url` varchar(128) NOT NULL,
                                  `def_resp_code` varchar(16) DEFAULT NULL,
                                  `def_resp_header` varchar(64) DEFAULT NULL,
                                  `def_resp_body` mediumtext,
                                  `owner` varchar(64) DEFAULT NULL,
                                  `desc` varchar(1024) DEFAULT NULL,
                                  `meta` varchar(1024) DEFAULT NULL,
                                  `status` varchar(128) NOT NULL,
                                  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                  PRIMARY KEY (`id`),
                                  UNIQUE KEY `url`(`url`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='interface';



CREATE TABLE `stub_rule` (
                             `id` int(32) NOT NULL AUTO_INCREMENT,
                             `interface_id` int(32) NOT NULL,
                             `match_type` int(32) NOT NULL COMMENT '1:match request query url, 2:match request body',
                             `match_rule` varchar(512) DEFAULT NULL,
                             `resp_code` varchar(16) DEFAULT NULL,
                             `resp_header` varchar(64) DEFAULT NULL,
                             `resp_body` mediumtext,
                             `delay_time` int(32) DEFAULT '0' COMMENT 'ms',
                             `desc` varchar(1024) DEFAULT NULL,
                             `meta` varchar(1024) DEFAULT NULL,
                             `status` varchar(128) NOT NULL,
                             `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`),
                             KEY `interface_id`(`interface_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='rule';