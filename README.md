# serverless-data-pipeline
SDCC project a.y. 2023/2024

## First step: build whole project
In no particular order:
 
 * Build the deployment program
 ~~~
 $ cd deploy
 $ bash build.nix.sh
 $ cd ..
 ~~~

 * Build the injector
 ~~~
 $ cd inject_data
 $ bash build.nix.sh
 $ cd ..
 ~~~

 * Build all the lambda packages
 ~~~
 $ cd lambdas
 $ bash build-pkg.nix.sh
 $ cd ..
 ~~~

Your Go distribuition will take care of getting all of the needed
dependencies.

A Windows version of the build scripts is also available: they are
build.win.bat or build-pkg.win.bat

## Second step: deploy the infrastructure

Prerequisite is to have an account with a role named "LabRole" with
enough permissions to create all the resources. If such a IAM role has
a different name on your account it can be changed in deploy/src/config.go
but requires recompilation of the deployment program.

Now, recover your AWS credentials (base64-encoded secret tokens) by visiting
AWS academy and downloading it by clicking on "AWS details" right after the
"Start lab"/"Stop lab" push buttons. A menu on the right will appear, which
shows AWS CLI: Show - just copy content of the secret token and paste into a
file on your computer called $HOME/.aws/credentials (or %USERPROFILE%/.aws/credentials if on Windows). Before doing that, you'll need to start the lab. This step has to be done on each restart of the lab, since security tokens
change every time.

In this particular order:

 1. Navigate to the deployment program built executable folder:
 ~~~
 $ cd deploy/bin
 ~~~

 2. Run the deployment program
 ~~~
 $ ./deploy
 ~~~

 3. Looking at the output log, take note of the API endpoint

 4. Navigate to the injector built executable folder:
 ~~~
 $ cd ../../inject_data/bin
 ~~~

 5. Launch the injector
 ~~~
 $ ./inject_data --api-endpoint <yourCopiedEndpoint> --every-ms 2000
 ~~~

NOTE: that on the first time it is being run, injector will download the dataset from my own Google Drive public folder

NOTE: limiting the rate by using --every-ms is strongly reccomended to avoid account deactivation (lots of lambdas running at the same time)

### Optional: enabling authentication

To enable authentication you may want to use deploy with -a option to set auth key to be able to trigger the preprocessing pipeline (undeploy if infrastructure already setup first, see next section):

~~~
$ ./deploy -a myownkey
~~~

And then, when you use the injector:

~~~
$ ./inject_data --auth-key myownkey --api-endpoint <yourCopiedEndpoint> --every-ms 2000
~~~

## Last step: undeployment

If you want to teardown the infrastructure:

1. Navigate to the deployment program built executable folder:
~~~
$ cd deploy/bin
~~~

2. Undeploy with -d option
~~~
$ ./deploy -d
~~~

NOTE: if you enabled authentication, the cryptographic storage containing the auth key, managed by "AWS Secret Manager" WILL NOT BE DELETED (on next deployments it will just be updated). This is because it will take 7 days for AWS to delete the encrypted storage, period in which this storage will not be usable and you will not be able to recreate a new crypto storage with the same name. However, maintaining whatever AWS resource takes some money, so if you want to delete it anyway use -s options along with -d while tearing down the infrastructure.
