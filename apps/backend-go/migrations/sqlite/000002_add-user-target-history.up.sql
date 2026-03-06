CREATE TABLE IF NOT EXISTS `user_target_history`
(
    `user_id` TEXT NOT NULL,
    `target_email` TEXT NOT NULL,
    `occurences` INT NOT NULL DEFAULT 1,
    `starred` INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(`user_id`, `target_email`),
    FOREIGN KEY(`user_id`) REFERENCES user(user_id)
);