package cmd

import (
	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

// 初始化所有的命令
type CommandService struct {
	commands map[common.CmdName]common.Cmd
}

func NewCommandService() *CommandService {
	return &CommandService{
		commands: make(map[common.CmdName]common.Cmd),
	}
}

func (s *CommandService) AddCommand(cmd common.Cmd) {
	if _, ok := s.commands[cmd.Name()]; ok {
		log.Errorf("command %s already exists", cmd.Name())
		return
	}
	s.commands[cmd.Name()] = cmd
}

func (s *CommandService) GetCommand(name common.CmdName) common.Cmd {
	if cmd, ok := s.commands[name]; ok {
		return cmd
	}
	return nil
}
