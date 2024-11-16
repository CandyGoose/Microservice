package main

import "context"

func NewBizModule() *BizModule {
	return &BizModule{}
}

type BizModule struct {
	UnimplementedBizServer
}

func (*BizModule) Check(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (*BizModule) Add(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (*BizModule) Test(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}
