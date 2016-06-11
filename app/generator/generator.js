'use strict';

angular.module('myApp.generator', ['ngRoute'])

    .config(['$routeProvider', function ($routeProvider) {
        $routeProvider.when('/generator', {
            templateUrl: 'generator/generator.html',
            controller: 'GeneratorCtrl'
        });
    }])

    .controller('GeneratorCtrl', [function () {

    }]);