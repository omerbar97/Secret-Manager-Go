## Secret Manager Go
The AWS Secret Manager CLI is a command-line tool that enables users to retrieve secrets from the AWS Secret Manager and display the access log in a CSV file or in the CLI itself.

### Overview
In this CLI application, I have implemented a multithreading approach to concurrently retrieve API results, utilizing the concept of rate limiting to ensure that the number of API calls does not exceed the imposed limit. I have effectively mitigated the risk of surpassing the API threshold set by AWS services. This strategy enables efficient and simultaneous data retrieval from the AWS services, optimizing the application's performance while ensuring adherence to the prescribed API limits, thereby preventing potential disruptions in service availability.

Additionally, the program incorporates a memory cache mechanism that not only facilitates swift access to recently retrieved information but also ensures data persistence. This memory cache is periodically stored as hard files in the server, serving as a reliable backup. Upon initiating the server, the application automatically loads the cached data from these files if they exist, enabling seamless continuity of operations. By employing this strategy, the application minimizes the reliance on repetitive API calls and reduces response time, thus optimizing the overall system performance. This approach not only enhances the speed of data retrieval but also significantly reduces the load on the system, contributing to a more streamlined and responsive user experience.

### Installation
Before using the CLI, ensure you have set up your AWS credentials. You can either set them up through the AWS CLI or set the 'public' and 'secret' environment variables inside the .env file (in the current working dir).

.env
```
public=
secret=
```

Cloning the project using git :
```
git clone https://github.com/omerbar97/Secret-Manager-Go.git
```

#### Without docker
```
go mod download                     -- Downloading dependencies
go run api/server/server.go         -- Starting the server
```
#### Using docker
Inside the project folder

If you have Make installed:
```
sudo make build                     -- Building the server image
sudo make run                       -- Starting the server image
```
without Make:
```
docker build -t sm .                -- Building the server image
docker run -p 8080:8080 sm          -- Starting the server image
```

```
go run cmd/cli/main.go              -- Starting the cli
```

to exit the program use ctrl + c

### Usage
The CLI offers the following functionalities:

#### Config Setup
```
>> load                            -- Loading the AWS keys from the .env file
>> load public <key>               -- Setting the AWS public key
>> load secret <key>               -- Setting the AWS secret key
>> load region <secret region>     -- Setting the AWS region 
```

#### Retrieving Secrets
```
the default zone is: "eu-north-1" you can change in inside the cli

>> get secrets                     -- Retriving all the user secret + metadata + access 
                                   log from the secret manager and saving it inside the 
                                   .csv file in the current folder  
```

#### Showing Reports
```
>> get report <secret_arn>         -- Displaying the report of each secret
```

### Requirements
```
- Git                       https://git-scm.com/downloads
- Go langauge               https://go.dev/doc/install
- Docker (optional)         https://docs.docker.com/get-docker/
- Make (optional)           https://www.gnu.org/software/make/
```

#### Auther
Omer Bar - Computer Science Student 3rd Year
