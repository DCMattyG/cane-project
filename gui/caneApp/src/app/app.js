app = angular.module('caneApp', ['ngStorage', 'ngRoute']);


app.config(function($routeProvider) {
	  $routeProvider

	  .when('/',	 {
		      templateUrl : 'home.html',
		      controller  : 'caneHomeController'
		    })

	  .when('/login', {
		      templateUrl : 'login.html',
		      controller  : 'caneLoginController'
		    })

	  .when('/signup', {
		      templateUrl : 'signup.html',
		      controller  : 'caneLoginController'
		    })

	  .when('/jobs', {
		      templateUrl : 'jobs.html',
		      controller  : 'caneJobsController'
		    })

	  .when('/workflows', {
		      templateUrl : 'workflows.html',
		      controller  : 'caneWorkflowController'
		    })

	  .when('/workflows/new', {
		      templateUrl : 'workflows_new.html',
		      controller  : 'caneWorkflowController'
		    })

	  .when('/deviceapis', {
		      templateUrl : 'deviceapis.html',
		      controller  : 'caneDeviceApiController'
		    })

	  .when('/deviceapis/new', {
		      templateUrl : 'deviceapis_new.html',
		      controller  : 'caneDeviceApiController'
		    })

	  .when('/devices', {
		      templateUrl : 'devices.html',
		      controller  : 'caneDevicesController'
		    })

	  .otherwise({redirectTo: '/'});
});



app.run(function($rootScope, AuthenticationService) {
	$rootScope.baseUrl = "http://cane.cisco.com/";
	$rootScope.token = "";
	
})


/**
 * @ngdoc type
 * @module caneApp
 * @name headerController
 *
 * @description
 *
 *
 */

app.controller('headerController', function($scope, $rootScope, $location, $http, $localStorage, AuthenticationService) {


	$rootScope.activeTab = "";
	$scope.userName = "";
	var updateUsername = function(){
		$scope.userName = AuthenticationService.authUserName;
	}

	AuthenticationService.registerObserverCallback(updateUsername);


	$scope.doLogout = function() {
		AuthenticationService.Logout();
		$location.path('/login');
		$rootScope.isAuthenticated = AuthenticationService.IsLoggedIn();
	}
});

/**
 * @ngdoc type
 * @module caneApp
 * @name caneHomeController
 *
 * @description
 *
 *
 */

app.controller('caneHomeController', function($scope, $rootScope, $location, $http, $localStorage, AuthenticationService) {

	$rootScope.isAuthenticated = AuthenticationService.IsLoggedIn();
	if (!$rootScope.isAuthenticated) {
		$location.path('/login');
	}

	$scope.userToken = $localStorage.currentUser.token;

});

/**
 * @ngdoc type
 * @module caneApp
 * @name caneLoginController
 *
 * @description
 *
 *
 */

app.controller('caneLoginController', function($scope, $rootScope, $location, $http, $location, $localStorage, AuthenticationService) {

	console.log("did load controller");

	$scope.showMessage = true;
	$scope.messageTitle = "Warning";
	$scope.messageText = "Do not use your CEC credentials at this time.";

	$scope.showErrorMessage = function (title, text) {
		$scope.messageTitle = title
		$scope.messageText = text;
		$scope.showMessage = true;
	}

	$scope.resetErrorMessage = function() {
		$scope.showMessage = false;
		$scope.messageTitle = "";
               	$scope.messageText = "";
	}

	$scope.doCaneLogin = function() {

		console.log("did click login");

		AuthenticationService.Login($scope.username, $scope.password, function (result) {
			if (result === true) {
				$http.defaults.headers.common.Authorization = 'Bearer ' +  $localStorage.currentUser.token;
				$location.path('/');
			} else {
				alert("Login Failed! Invalid username or password");
			}
		});

	};

	$scope.doCaneCreateAccount = function() {

                var data = {
                        fname: $scope.fname,
                        lname: $scope.lname,
                        username: $scope.username,
                        password: $scope.password,
			privilege: 1,
			enable: true
                };
		data = JSON.stringify(data);

		AuthenticationService.RegisterNewAccount(data, function (result) {
			if (result === true) {
				$http.defaults.headers.common.Authorization = 'Bearer ' +  $localStorage.currentUser.token;
				$location.path('/');
			} else {
				alert("Registration Failed!");
			}
		});
        };
});

/**
 * @ngdoc type
 * @module caneApp
 * @name caneJobsController
 *
 * @description
 *
 *
 */

app.controller('caneJobsController', function($scope, $rootScope, $location, $interval, $http, $localStorage, AuthenticationService) {

	$rootScope.isAuthenticated = AuthenticationService.IsLoggedIn();
	if (!$rootScope.isAuthenticated) {
		$location.path('/login');
	}
	$http.defaults.headers.common.Authorization = 'Bearer ' +  $localStorage.currentUser.token;


	$scope.showMessage = false;
	$scope.messageTitle = "";
	$scope.messageText = "";

	$scope.showErrorMessage = function (title, text) {
		$scope.messageTitle = title
		$scope.messageText = text;
		$scope.showMessage = true;
	}
	$scope.resetErrorMessage = function() {
		$scope.showMessage = false;
		$scope.messageTitle = "";
		$scope.messageText = "";
	}

	$scope.jobsList = [];
	$scope.jobsObject = {};
	$scope.jobsInProgress = [];

        $scope.getJobs = function() {

                $http({
                        url: $rootScope.baseUrl + 'claim',
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
				angular.forEach(response.data.claims, function(claimCode, key) {
					$scope.getClaimDetail(claimCode);
				});

                        }, function myError(response) {
                        });
        }
        $scope.getClaimDetail = function(claimCode) {
                $http({
                        url: $rootScope.baseUrl + 'claim' + "/" + claimCode,
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
				$scope.jobsList.push(response.data);

				var jobStatus = response.data.currentStatus;
				if (jobStatus == 0 || jobStatus == 1) {
					$scope.jobsInProgress.push(claimCode);
				}
				var timestamp = response.data.timestamp;
				timestamp = timestamp.split(" ");
				timestamp = timestamp[1].split(".");
				timestamp = timestamp[0];
				response.data.timestampHuman = timestamp;
				$scope.jobsObject[claimCode] = response.data;

				console.log($scope.jobsObject);

                        }, function myError(response) {
                                //alert("There was an error");
                        });
        }

	var checkJobs = function() {
		console.log("checking jobs");
		angular.forEach($scope.jobsInProgress, function(claimCode, key) {
			$scope.getClaimDetail(claimCode);
		});
	}
	$interval(checkJobs, 5000);

	$scope.getJobs();

});


/**
 * @ngdoc type
 * @module caneApp
 * @name caneDevicesController
 *
 * @description
 *
 *
 */

app.controller('caneDevicesController', function($scope, $rootScope, $location, $http, $localStorage, AuthenticationService) {

	$rootScope.isAuthenticated = AuthenticationService.IsLoggedIn();
	if (!$rootScope.isAuthenticated) {
		$location.path('/login');
	}
	$http.defaults.headers.common.Authorization = 'Bearer ' +  $localStorage.currentUser.token;


	$scope.showMessage = false;
	$scope.messageTitle = "";
	$scope.messageText = "";

	$scope.showErrorMessage = function (title, text) {
		$scope.messageTitle = title
		$scope.messageText = text;
		$scope.showMessage = true;
	}
	$scope.resetErrorMessage = function() {
		$scope.showMessage = false;
		$scope.messageTitle = "";
		$scope.messageText = "";
	}


	$scope.myDevices = [];

	/*********************************
	 * Get list of Cane devices
	 * ******************************/
	$scope.getCaneDevices = function() {

		$scope.myDevices = [];
                $scope.myDeviceList = {};
		$scope.resetErrorMessage();
		$http({
                        url: $rootScope.baseUrl + 'device',
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
                                $scope.myDeviceList = response.data.devices;

				angular.forEach($scope.myDeviceList, function(tempDeviceName, key) {
					$scope.getCaneDeviceDetail(tempDeviceName);
				});

                        }, function myError(response) {
				$scope.showErrorMessage("Oh no", "There was an error retreiving devices from the server");
                        });
	}

	/*********************************
	 * Get Cane device detail
	 * ******************************/
	$scope.getCaneDeviceDetail = function(deviceName) {

		$scope.resetErrorMessage();
                $http({
                        url: $rootScope.baseUrl + 'device/' + deviceName,
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
                                //console.log(response);
				/*if (!response.data.devicetype || response.data.devicetype == 0) {
					response.data.devicetype = "Generic";
					response.data.libUrl = "https://www.google.com";
				}
				else if (response.data.devicetype = 1) {
					response.data.devicetype = "Meraki";
					response.data.libUrl = "";
				}
				else if (response.data.devicetype = 2) {
					response.data.devicetype = "DNA Center";
					response.data.libUrl = "";
				}
				else if (response.data.devicetype = 3) {
					response.data.devicetype = "Intersight";
					response.data.libUrl = "";
				}
				else if (response.data.devicetype = 4) {
					response.data.devicetype = "UCS Director";
					response.data.libUrl = "";
				}
				else if (response.data.devicetype = 5) {
					response.data.devicetype = "Prime Infrastructure";
					response.data.libUrl = "";
				}*/

				if (!response.data.lastRefresh) {
					response.data.lastRefresh = "Unknown";
				}

				response.data.endpoint = $rootScope.baseUrl + response.data.device.name + "/";
				response.data.Status = 0;
                                $scope.myDevices.push(response.data);
				//console.log($scope.myDevices);
                        }, function myError(response) {
				//alert("There was an error");
				$scope.showErrorMessage("Oh no", "There was an error retreiving device details from the server");
                        });
        }

	$scope.resetDeviceAddForm = function() {

		// clear inputs in modal
		$scope.name = "";
		$scope.ipaddress = "";
		$scope.username = "";
		$scope.password = "";
		$scope.authtype = "";
	}

	/*********************************
	 * Add a new Cane Device
	 * ******************************/
	$scope.addCaneDevice = function() {

		var data = {
			device: {
				name: $scope.name,
				url: $scope.ipaddress,
				authType: $scope.authtype
			},
			auth: {}
		};

		if ($scope.authtype == "basic") {
			data.auth.username = $scope.basic_username;
			data.auth.password = $scope.basic_password;
		}
		else if ($scope.authtype == "apikey") {
			data.auth.header = $scope.apikey_header;
			data.auth.key = $scope.apikey_key;
		}
		else if ($scope.authtype == "session") {
			data.auth.username = $scope.session_username;
			data.auth.password = $scope.session_password;
			data.auth.authBody = $scope.session_authbody;
			data.auth.authBodyMap = $scope.session_authbodymap;
			data.auth.cookieLifetime = $scope.session_lifetime;
		}
		else if ($scope.authtype == "rfc3447") {
			data.auth.publicKey = $scope.rfc3447_public;
			data.auth.privateKey = $scope.rfc3447_private;
		}


                $http({
                        url: $rootScope.baseUrl + 'device',
                        dataType: 'json',
                        method: 'POST',
                        data: JSON.stringify(data),
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
				$scope.getCaneDevices();

                        }, function myError(response) {
				alert("There was an error adding the device");
                        });

		$scope.resetDeviceAddForm();
	}

	$scope.openDeviceModal = function(deviceId) {
		var modalId = "modal-editDevice-" + deviceId;
		console.log("trying to open " + modalId);
		openModal(modalId);
	}

	$scope.getCaneDevices();

});


/**
 * @ngdoc type
 * @module caneApp
 * @name caneWorkflowController
 *
 * @description
 *
 *
 */

app.controller('caneWorkflowController', function($scope, $rootScope, $location, $http, $localStorage, AuthenticationService) {

	$rootScope.isAuthenticated = AuthenticationService.IsLoggedIn();
	if (!$rootScope.isAuthenticated) {
		$location.path('/login');
	}
	$http.defaults.headers.common.Authorization = 'Bearer ' +  $localStorage.currentUser.token;


	$scope.showMessage = false;
	$scope.messageTitle = "";
	$scope.messageText = "";

	$scope.showErrorMessage = function (title, text) {
		$scope.messageTitle = title
		$scope.messageText = text;
		$scope.showMessage = true;
	}
	$scope.resetErrorMessage = function() {
		$scope.showMessage = false;
		$scope.messageTitle = "";
		$scope.messageText = "";
	}

	$scope.workFlowList = [];
	$scope.getWorkflows = function() {

                $http({
                        url: $rootScope.baseUrl + 'workflow',
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
				angular.forEach(response.data.workflows, function(tempDeviceName, key) {
					$scope.getWorkflowDetails(tempDeviceName);
				});

                        }, function myError(response) {
				alert("There was an error");
                        });
        }

	$scope.getWorkflowDetails = function(name) {

                $http({
                        url: $rootScope.baseUrl + 'workflow' + "/" + name,
                        dataType: 'json',
                        method: 'GET'
                }).then (
                        function mySuccess(response) {
				$scope.workFlowList = $scope.workFlowList.concat(response.data);

                        }, function myError(response) {
				alert("There was an error");
                        });
	}

	$scope.getWorkflows();

	$scope.myDeviceList = [];
	$scope.myDeviceApis = {};
	$scope.workflowSteps = [];

	$scope.addStep = function() {
		var keys = Object.keys($scope.workflowSteps);
		var len = keys.length;

		var numSteps = $scope.workflowSteps.length;
		var newStepNumber = len;
		var stepObject = {
			stepID: newStepNumber,
			deviceName: "",
			deviceAPI: "",
			deviceMap: "",
			input1: "step_" + newStepNumber + "_device",
			input2: "step_" + newStepNumber + "_api",
			input3: ""
		};
		
		$scope.workflowSteps.push(stepObject);
	}


        /*********************************
         * Get list of Cane devices
         * ******************************/
        $scope.getCaneDevices = function() {

		$scope.workFlowList = [];
                $http({
                        url: $rootScope.baseUrl + 'device',
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json",
                                "Authorization": 'Bearer ' +  $localStorage.currentUser.token
                        }
                }).then (
                        function mySuccess(response) {
                                $scope.myDeviceList = response.data.devices;

				angular.forEach($scope.myDeviceList, function(tempDeviceName, key) {
					$scope.getApisForDevice(tempDeviceName);
				});

                        }, function myError(response) {
				alert("There was an error");
                        });
        }

	$scope.getApisForDevice = function(deviceName) {

		 $http({
                        url: $rootScope.baseUrl + 'api' + "/" + deviceName,
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json",
                                "Authorization": 'Bearer ' +  $localStorage.currentUser.token
                        }
                }).then (
                        function mySuccess(response) {
				$scope.myDeviceApis[deviceName] = response.data.apis;

                        }, function myError(response) {
                        });
	}

	$scope.saveWorkflow = function() {

		console.log("going to save workflow");
		console.log($scope.workflowSteps);

		var data = {
			name: $scope.workflowName,
			type: $scope.workflowCategory,
			description: $scope.workflowDescription,
			steps: []
		};

		angular.forEach($scope.workflowSteps, function(tempObject, key) {
			
			var parsedMap = [];
			var lines = tempObject.input3.split('\n');
			for(var i = 0;i < lines.length;i++){
				var line = lines[i].split('=');
				var parsedLine = {};
				parsedLine[line[0]] = line[1];
				parsedMap.push(parsedLine);
			}
			console.log("map is " + parsedMap);

			var stepData = {
				stepNum: tempObject.stepID,
				deviceAccount: tempObject.input1,
				apiCall: tempObject.input2,
				varMap: parsedMap	
			};
			data.steps.push(stepData);

		});

		console.log(data);

		$http({
                       	url: $rootScope.baseUrl + 'workflow',
                       	dataType: 'json',
                       	method: 'POST',
                       	data: JSON.stringify(data),
                       	headers: {
                               	"Content-Type": "application/json"
                        	}
                }).then (
                       	function mySuccess(response) {
				$location.path("/workflows");
	
                       	}, function myError(response) {
				alert("There was an error");
				console.log(response);
                });

	}

	$scope.deleteWorkflow = function(workflowName) {

                $http({
                        url: $rootScope.baseUrl + 'workflow' + "/" + workflowName,
                        dataType: 'json',
                        method: 'DELETE'
                }).then (
                        function mySuccess(response) {
                                $scope.getCaneDevices();
                        }, function myError(response) {
                                alert("Delete api error!: " + response.data.message);
                        });
        }

	$scope.startWorkflow = function(name) {
		$http({
                       	url: $rootScope.baseUrl + 'workflow' + "/" + name,
                       	method: 'POST'
                }).then (
                       	function mySuccess(response) {
				console.log(response);
				if (response.data.claimCode) {
					alert("Worklow has started! Claim code is " + response.data.claimCode);
				}
	
                       	}, function myError(response) {
				alert("There was an error");
				console.log(response);
                });

	}

	$scope.getCaneDevices();

});

/**
 * @ngdoc type
 * @module caneApp
 * @name caneDeviceApiController
 *
 * @description
 *
 *
 */

app.controller('caneDeviceApiController', function($scope, $rootScope, $location, $window, $http, $localStorage, AuthenticationService) {

	$rootScope.isAuthenticated = AuthenticationService.IsLoggedIn();
	if (!$rootScope.isAuthenticated) {
		$location.path('/login');
	}
	$http.defaults.headers.common.Authorization = 'Bearer ' +  $localStorage.currentUser.token;



	$scope.showMessage = false;
	$scope.messageTitle = "";
	$scope.messageText = "";

	$scope.showErrorMessage = function (title, text) {
		$scope.messageTitle = title
		$scope.messageText = text;
		$scope.showMessage = true;
	}
	$scope.resetErrorMessage = function() {
		$scope.showMessage = false;
		$scope.messageTitle = "";
		$scope.messageText = "";
	}

	 $scope.myDeviceList = [];
         $scope.myDeviceApis = {};
         $scope.myDeviceApisFlat = [];

        /*********************************
         * Get list of Cane devices
         * ******************************/
        $scope.getCaneDevices = function() {

	 	$scope.myDeviceList = [];
         	$scope.myDeviceApis = {};
         	$scope.myDeviceApisFlat = [];
		$scope.resetErrorMessage();
                $http({
                        url: $rootScope.baseUrl + 'device',
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
                                $scope.myDeviceList = response.data.devices;

                                angular.forEach($scope.myDeviceList, function(tempDeviceName, key) {
                                        $scope.getApisForDevice(tempDeviceName);
                                });

                        }, function myError(response) {
				$scope.showErrorMessage("Oh no", "There was an error retreiving devices from the server");
                        });
        }

        $scope.getApisForDevice = function(deviceName) {

		$scope.resetErrorMessage();
                $http({
                        url: $rootScope.baseUrl + 'api' + "/" + deviceName,
                        dataType: 'json',
                        method: 'GET',
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
                               	$scope.myDeviceApis[deviceName] = response.data.apis;

				if (response.data.apis != null && response.data.apis.length > 0) {
					angular.forEach(response.data.apis, function(apiname, key) {
						$scope.getApiDetails(deviceName, apiname);
					});
				}

                        }, function myError(response) {
				$scope.showErrorMessage("Oh no", "There was an error retreiving device apis from the server");
                        });
        }

	$scope.getApiDetails = function(deviceName, apiName) {

                $http({
                        url: $rootScope.baseUrl + 'api' + "/" + deviceName + "/" + apiName,
                        dataType: 'json',
                        method: 'GET'
                }).then (
                        function mySuccess(response) {
				$scope.myDeviceApisFlat.push(response.data);	
				console.log($scope.myDeviceApisFlat);

                        }, function myError(response) {
				$scope.showErrorMessage("Oh no", "There was an error retreiving device apis from the server");
                        });
	}

	$scope.parseResponse = function() {

                $http({
                        url: $rootScope.baseUrl + 'parseVars',
                        method: 'POST',
                        dataType: 'json',
                        headers: {
                                "Content-Type": "application/json"
			},
                        data: $scope.inputParams
                }).then (
                        function mySuccess(response) {
				$scope.inputParams = response.data.parsedAPI;

                        }, function myError(response) {
				alert("There was an error");
                        });

        }

	$scope.deleteAPI = function(device, api) {

         	$http({
                        url: $rootScope.baseUrl + 'api' + "/" + device + "/" + api,
                        dataType: 'json',
                        method: 'DELETE'
                }).then (
                        function mySuccess(response) {
				$scope.getCaneDevices();
                        }, function myError(response) {
				alert("Delete api error!: " + response.data.message);
                        });
	}

	$scope.saveAPI = function() {
		var data = {
                        name: $scope.name,
                        deviceAccount: $scope.devicename,
                        url: $scope.apiendpoint,
                        body: $scope.inputParams,
                        method: $scope.method,
                        type: $scope.datatype
                };

         	$http({
                        url: $rootScope.baseUrl + 'api',
                        dataType: 'json',
                        method: 'POST',
                        data: JSON.stringify(data),
                        headers: {
                                "Content-Type": "application/json"
                        }
                }).then (
                        function mySuccess(response) {
				console.log("API created");
				$location.path("/deviceapis");
                        }, function myError(response) {
				alert("add api error!: " + response.data.message);
                        });
	}


	$scope.getCaneDevices();
});
