function getUrlParameter(name) {
    name = name.replace(/[\[]/, '\\[').replace(/[\]]/, '\\]');
    var regex = new RegExp('[\\?&]' + name + '=([^&#]*)');
    var results = regex.exec(location.search);
    return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
};

app.controller('Account', [
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

        Ape.Request("GET", "/account_status", {}, (data) => {
            if (data) {
                $scope.hasCard = true;
            }
        })

        $scope.updatePassword = () => {


            Ape.Request("POST", "/update_password", {
                Email: sha256($scope.data.cpassword),
                Password: sha256($scope.data.npassword)
            }, (data) => {

                if (!data) {
                    swal("Error", "Incorrect current password entered.", "error")
                    return;
                }

                swal("Success", "Password updated!", "success")
            })
        }

        $scope.deleteAccount = () => {
            swal({
                    title: "Are you sure?",
                    text: "Once deleted, you will not be able to recover your account?",
                    icon: "warning",
                    buttons: true,
                    dangerMode: true,
                })
                .then((willDelete) => {
                    if (willDelete) {
                        Ape.Request("GET", "/delete_account", {}, (data) => {
                            if (data)
                                window.location = "/login";
                            else swal("Error", "Your account was not deleted, please try again", "error")
                        })
                    }
                });
        }

        var success = getUrlParameter("success"),
            error = getUrlParameter("error");

        if (success) {
            swal("Success", success, "success");
        }

        if (error) {
            swal("Error", error, "error");
        }

    }
]);