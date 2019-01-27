app.service('AuthenticationService', function($http, $localStorage) {

	console.log("Authentication Service loaded");

	this.authUserName = "";

	var observerCallbacks = [];

  	//register an observer
  	this.registerObserverCallback = function(callback){
		console.log("callback registered");
    		observerCallbacks.push(callback);
  	};
	
  	//call this when you know 'foo' has been changed
  	var notifyObservers = function(){
    		angular.forEach(observerCallbacks, function(callback){
			console.log("calling callback function");
      			callback();
    		});
  	};




	this.getUsername = function() {
		return this.authUserName;
	}

	var saveAuthData = function(var_username, var_token) {
        	// store username and token in local storage to keep user logged in between page refreshes
        	$localStorage.currentUser = { username: var_username, token: var_token };
        	// add jwt token to auth header for all requests made by the $http service
        	$http.defaults.headers.common.Authorization = 'Bearer ' + var_token;

		return true;
	}

	this.RegisterNewAccount = function(jsonData) {
		// make sure there is no local data for the app
            	delete $localStorage.currentUser;
            	$http.defaults.headers.common.Authorization = '';

            	$http.post('http://cane.cisco.com/user', jsonData)
                	.success(function (response) {
                    	// login successful if there's a token in the response
                    	if (response.token) {
				console.log("login successfull");
				this.saveAuthData(response.username, response.token);
                        	// execute callback with true to indicate successful login
                        	callback(true);
                    	} else {
				console.log("login failed");
                        	// execute callback with false to indicate failed login
                        	callback(false);
                    	}
                	}).error( function(response) {
				console.log("auth error");
				callback(false);
			});

	}

        this.Login = function (username, password, callback) {

            // remove user from local storage and clear http auth header
            delete $localStorage.currentUser;
            $http.defaults.headers.common.Authorization = '';

            $http.post('http://cane.cisco.com/login', { username: username, password: password })
                .success(function (response) {
                    // login successful if there's a token in the response
                    if (response.token) {

			console.log("login successfull");
			saveAuthData(username, response.token);
                        // execute callback with true to indicate successful login
                        callback(true);
                    } else {
			console.log("login failed");
                        // execute callback with false to indicate failed login
                        callback(false);
                    }
                }).error( function(response) {
			console.log("auth error");
			callback(false);
		});
        }
 
        this.Logout = function() {
		console.log("Logging out");
		
            	// remove user from local storage and clear http auth header
            	delete $localStorage.currentUser;
            	$http.defaults.headers.common.Authorization = '';
		this.authUserName = "";
		notifyObservers();
        }

        this.IsLoggedIn = function() {
		if ($localStorage.currentUser && $localStorage.currentUser.token) {
			this.authUserName = $localStorage.currentUser.username;
		    	notifyObservers();
			return true;
		}
		else {
			this.authUserName = "";
		    	notifyObservers();
			return false;
		}
	}

});
