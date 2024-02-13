# goswag
To help you auto generate swagger for your golang code.

<p align="center">
    <a href="https://github.com/diegoclair/goswag/tags" alt="GitHub tag">
     <img src="https://img.shields.io/github/tag/diegoclair/goswag.svg" />
    </a>
    <a href='https://coveralls.io/github/diegoclair/goswag?branch=main'>
     <img src='https://coveralls.io/repos/github/diegoclair/goswag/badge.svg?branch=main' alt='Coverage Status' />
    </a>
    <a href="https://github.com/diegoclair/goswag/actions">
     <img src="https://github.com/diegoclair/go_utils/actions/workflows/ci.yaml/badge.svg" alt="build status">
    </a>
    <a href="https://github.com/diegoclair/goswag/contributors" alt="Contributors">
     <img src="https://img.shields.io/github/contributors/diegoclair/goswag" />
    </a>
    <a href="https://opensource.org/licenses/MIT">
     <img src="https://img.shields.io/badge/License-MIT-yellow.svg" />
    </a>
</p>


### Setup
- Create a folder on your repository root named goswag.  
- Create a file called main.go with package main
- Add setup for your router
- create a command on make file

## Contributing

Contributions are welcomed. To contribute, please follow these steps:

1. Fork the repository
2. Create a new feature branch (`git checkout -b feature/<FEATURE NAME>`)
3. Make the necessary changes
4. Commit your changes (`git commit -m "Add some feature"`)
5. Push your changes to your forked repository (`git push origin feature/<FEATURE NAME>`)
6. Create a pull request to the main branch of the repository

## License

Goswag is [MIT licensed](./LICENSE).

### TODO:
- Create ReadmeDocs
- Gin does not implement the method Match or Any from gin, because there are no way (yet) to define the same (summary,responses,bodies) for all methods
