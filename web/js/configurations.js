app.controller('Configurations', [
    '$scope',
    'Ape',
    function($scope, Ape) {

        var editor;

        Ape.Init({
            start: () => {
                $scope.loading = true
            },
            end: () => {
                $scope.loading = false
            },
            base: ""
        })

        $scope.current = {}

        $scope.getConfigs = () => {
            Ape.Request("GET", "/configurations", {}, (data, response) => {
                if (!data && response.status != 200) {
                    alert("Your session has expired, please signin again.")
                    window.location = "/login";
                    return;
                }
                $scope.items = data;
            })
        }

        $scope.remove = (config) => {

            if (config.Nodes > 0) {
                swal("Error", "You must poweroff all of the active instances using this configuation.", "error")
                return;
            }

            swal({
                    title: "Are you sure?",
                    text: "Once deleted, you will not be able to recover this resource.",
                    icon: "warning",
                    buttons: true,
                    dangerMode: true,
                })
                .then((willDelete) => {
                    if (willDelete) {
                        Ape.Request("DELETE", "/configurations", { id: config.Id }, (data) => {
                            if (!data) {
                                swal("Error", "error removing configuation, please try again", "error")
                                return
                            }

                            var index = $scope.items.indexOf(config);
                            $scope.items.splice(index, 1);
                            $scope.current = {};
                            swal("Success", "Configuration removed.", "success")
                        })
                    }
                });
        }

        $scope.modal = () => {
            $('.modal').modal('show');
            $(".code-target").html($scope.current.ApiKey);
        }


        $scope.open = (config) => {
            $scope.loading = true;
            var current = config;
           

            pasync(() => {
                $scope.current = current;
                $scope.loading = false;
                $scope.$apply();

                $(function() {
                    $('#configTabs  a').click((e) => {
                        e.preventDefault();

                        $(e.target).tab('show');
                    })
                })

                pasync(() => {
                    var container = document.getElementById("jsoneditor");
                    var options = {
                        mode: 'tree'
                    };
                    editor = new JSONEditor(container, options);
                    editor.set(config.Config);
                })

            })

        }

        $scope.close = () => {
            $scope.current = {};
        }

        $scope.add = () => {
            Ape.Request("POST", "/configurations", {}, (data) => {
                if (!data) {
                    swal("Error", "error creating configuation, please try again", "error")
                    return
                }

                swal("Success", "Configuration created.", "success")
                $scope.getConfigs();
            })
        }

        $scope.update = (config) => {
            config.Config = editor.get();
            Ape.Request("PUT", "/configurations", config, (data) => {
                if (!data) {
                    swal("Error", "error updating configuation, please try again", "error")
                    return
                } 

                swal("Success", "Configuration saved.", "success")

            })
        }

        $scope.rollConfiguration

        $scope.bash = {};
        $scope.addCommand = () => {
            $scope.current.Commands.push($scope.bash.cmd);
        }

        $scope.removeCommand = (index) => {
            $scope.current.Commands.splice(index, 1);
        }

        $scope.getConfigs();
    }
]);