#!/bin/bash

migrate -source file://migrations/ -database mysql://root:password@/srcabl_users up
