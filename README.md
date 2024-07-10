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

Prerequisite is to have an account with a role named "LabRole" 
(instructure AWS academy's default IAM role, at least for students, 
with no possibility to create another one) with
enough permissions to create all the resources. If such a IAM role has
a different name on your account it can be changed in deploy/src/config.go
(change the constant IAM_ROLE, line 19).

You will also need to be able to create resources in "us-east-1" 
(this is one of the few regions if not the ONLY region that can be 
used with instructure's AWS academy) AWS region,
as explained above, if this is not the case, the region can be changed in
deploy/src/config.go (change the constant AWS_REGION, line 18).

Requires recompilation of the deployment program 
(just redo the first point of the previous section after changes to the constant)

Now, recover your AWS credentials (base64-encoded secret tokens) by launching
AWS academy and copying them by clicking on "AWS details" right after the
"Start lab"/"Stop lab" push buttons. A menu on the right will appear, which
shows a button AWS CLI: Show - click it - then just copy shown content and paste into a
file on your computer called $HOME/.aws/credentials 
(or %USERPROFILE%/.aws/credentials if on Windows). 
Before doing this, you'll need to start the lab. 
This step has to be done on each restart of the lab, since security tokens
change every time AWS lab is started.

In this particular order (starting from this project root):

 1. Navigate to the deployment program built executable folder:
 ~~~
 $ cd deploy/bin
 ~~~

 2. Run the deployment program
 ~~~
 $ ./deploy
 ~~~

 3. Looking at the output log, take note/copy of the API endpoint

 4. Navigate to the injector built executable folder:
 ~~~
 $ cd ../../inject_data/bin
 ~~~

 5. Launch the injector
 ~~~
 $ ./inject_data --api-endpoint <yourCopiedEndpoint> --every-ms 2000
 ~~~

The injector will start to read the dataset (dirtying tuples, if it chose to, based on values got by a PRNG) and push tuples to the preprocessing pipeline!!

The injector can be stopped using CTRL+C, since some data have already been sent to the pipeline: when restarting the injector, just use --start-at option to define at which line
of the dataset the injector must resume sending tuples.

NOTE: that on the first time it is being run, injector will download the dataset from my own Google Drive public folder

NOTE: limiting the HTTP request issuing rate by using --every-ms is **strongly** reccomended to avoid account deactivation (lots of lambdas running at the same time)

### AWS console: see results

After running the injector, access your own AWS web console and see results of executing the step function 
called "CriticalDataPipeline" and relative DynamoDB tables "validationStatus", "transformationStatus", 
"storeStatus" and the final one (to be queried by an hypothetical data processing client/consumer) which is "nycYellowTaxis"

### Optional: enabling authentication

When deploying, you may want to use the deploy program with -a option to enable authentication and requiring auth key
to be able to trigger the preprocessing pipeline (undeploy if infrastructure already setup first, see next section):

~~~
$ ./deploy -a myownkey
~~~

And then, when you use the injector specifying --auth-key <authKey>:

~~~
$ ./inject_data --auth-key myownkey --api-endpoint <yourCopiedEndpoint> --every-ms 2000
~~~

## Last step: undeployment

If you want to teardown the infrastructure (starting from this project root, you should ensure having valid credentials file):

1. Navigate to the deployment program built executable folder:
~~~
$ cd deploy/bin
~~~

2. Undeploy with -d option
~~~
$ ./deploy -d
~~~

NOTE: if you enabled authentication, the cryptographic storage containing the auth key, 
managed by "AWS Secret Manager" WILL NOT BE DELETED (on next deployments it will just be updated). 
This is because it will take 7 days for AWS to delete the encrypted storage, 
period in which this storage will not be usable and you will not be able to recreate 
a new crypto storage with the same name. However, maintaining whatever AWS resource takes some money, 
so if you want to delete it anyway use -s options along with -d while passing cmdline options to deployment program.
