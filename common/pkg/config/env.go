package config

import (
	"errors"
	"os"
)

type Environment string

const (
	EnvTesting    = Environment("testing")    // 测试环境
	EnvProduction = Environment("production") // 生产环境
)

// 实现fmt.Stringer接口，调用方法获取其字符串
func (env *Environment) String() string {
	return string(*env)
}

func (env *Environment) Testing() Environment {
	return EnvTesting
}

func (env *Environment) Production() Environment {
	return EnvProduction
}

// 检查环境变量是否有效
func (env Environment) Invalid() bool {
	return env != EnvTesting && env != EnvProduction
}

// 读取全局配置的环境变量
func NewGlobalEnvironment() (Environment, error) {
	environment, ok := os.LookupEnv("ENVIRONMENT")
	if !ok {
		return "", errors.New("system environment:ENVIRONMENT not found")
	}

	env := Environment(environment)
	if env != EnvTesting && env != EnvProduction {
		return "", errors.New("environment not support, must be production, development")
	}

	return env, nil
}
