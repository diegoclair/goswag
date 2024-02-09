# goswag
To help you auto generate swagger for your golang code.

<p align="center">
    <a href="https://github.com/diegoclair/goswag/contributors" alt="Contributors">
        <img src="https://img.shields.io/github/contributors/diegoclair/goswag" /></a>
    <a href="https://github.com/diegoclair/goswag/pulse" alt="Activity">
        <img src="https://img.shields.io/github/commit-activity/m/diegoclair/goswag" /></a>
    <a href="https://github.com/diegoclair/goswag/actions">
        <img src="https://github.com/diegoclair/go_utils/actions/workflows/ci.yaml/badge.svg" alt="build status"></a>
</p>


### Setup
- Create a folder on your repository root named goswag.  
- Create a file called main.go with package main
- Add setup for your router
- create a command on make file

### TODO:
- Create ReadmeDocs
- Unit Tests
- Gin does not implement the method Match or Any from gin, because there are no way (yet) to define the same (summary,responses,bodies) for all methods
