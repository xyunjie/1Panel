package db_conf

import _ "embed"

//go:embed mysql_docker_compose.yml
var MysqlDockerCompose []byte

//go:embed mysql_env
var MysqlEnv []byte

//go:embed postgresql_docker_compose.yml
var PostgresqlDockerCompose []byte

//go:embed postgresql_env
var PostgresqlEnv []byte

//go:embed redis_docker_compose.yml
var RedisDockerCompose []byte

//go:embed redis_env
var RedisEnv []byte
