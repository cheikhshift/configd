 window.app = angular.module('app', []);

 app.factory('Ape', function($http) {

     var _Ape = { start: function() {}, base: "", headers: {}, end: function(response) {} };
     //configs ,  
     var StartCbs = [],
         EndCbs = [];

     var urlReq = function(verb, url, data, callback) {
         var lastError = {};

         for (var i = StartCbs.length - 1; i >= 0; i--) {
             StartCbs[i]();
         }

         lastError.params = ((verb != "POST" && verb != "PUT" && Object.keys(data).length > 0) ? `?${serialize(data, "")}` : ``)

         $http({
             method: verb,
             url: `${_Ape.base}${url}${lastError.params}`,
             data: data,
             headers: _Ape.headers
         }).then(function successCallback(response) {
             // this callback will be called asynchronously

             // when the response is available
             //  console.log(response);
             for (var i = EndCbs.length - 1; i >= 0; i--) {
                 EndCbs[i](response.data, response)
             }

             callback(response.data, response);

         }, function errorCallback(response) {
             // called asynchronously if an error occurs
             // or server returns response with an error status.
             //  $(".loader-overlay").css('display', 'none');

             for (var i = EndCbs.length - 1; i >= 0; i--) {
                 EndCbs[i](response.data, response)
             }
             callback(false, response);

         });
     };

     return {
         Init: function(Ape) {
             StartCbs.push(Ape.start);
             EndCbs.push(Ape.end);
             _Ape = Ape;
         },
         LoopEq: function(arr, key, value) {
             for (var i = arr.length - 1; i >= 0; i--) {
                 if (arr[i][key] == value) {
                     return arr[i]
                 }
             };
             return false;
         },
         Request: function(method, url, data, callback) {
             urlReq(method, url, data, callback);
         },
         getSelf: function() {
             return _Ape;
         }
     }
 });

 window.serialize = (obj, prefix) => {
     var str = [],
         p;
     for (p in obj) {
         if (obj.hasOwnProperty(p)) {
             var k = prefix ? prefix + "[" + p + "]" : p,
                 v = obj[p];
             str.push((v !== null && typeof v === "object") ?
                 serialize(v, k) :
                 encodeURIComponent(k) + "=" + encodeURIComponent(v));
         }
     }
     return str.join("&");
 };

 window.pasync = (cb) => {
 	window.setTimeout(() => {
 		cb()
 	}, 700);
 }
