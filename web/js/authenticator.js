app.controller('Authenticator', [
    '$scope',
    'Ape',
    function($scope, Ape) {
        $scope.data = {};

        Ape.Init({
            start: () => {
                $scope.loading = true
            },
            end: () => {
                $scope.loading = false
            },
            base: ""
        })

        Ape.Request("GET", "/logout", {}, () => {});

        $scope.login = () => {
            Ape.Request(
                "POST",
                "/login", {
                    Email: $scope.data.Email,
                    Password: sha256($scope.data.Password)
                },
                (data) => {
                    if (!data) {
                        $scope.authError("Invalid email/password combination.")
                        return;
                    }

                    window.location = "/admin";
                })
        }

        $scope.join = () => {
            Ape.Request(
                "POST",
                "/join", {
                    Email: $scope.data.Email,
                    Password: sha256($scope.data.Password)
                },
                (data) => {
                    if (!data) {
                        $scope.authError("Email in use already!")
                        return;
                    }

                    window.location = "/admin";
                })
        }

        $scope.reset = () => {

            Ape.Request("POST", "/reset", $scope.data, (data) => {
                if (!data) {
                    $scope.authError("Account specified was not found!")
                    return;
                }
                $scope.authSuccess("Please check your email for your new password.");
            })
        }

        $scope.authError = (str) => {
            swal("Error", str, "error")
        }

        $scope.authSuccess = (str) => {
            swal("Success", str, "success")
        }
    }
]);