'use strict'

function prettyDate(date) {
  return [
    date.getUTCFullYear(),
    '-',
    (date.getUTCMonth() < 9 ? '0' : '') + (date.getUTCMonth() + 1),
    '-',
    (date.getUTCDate() < 10 ? '0' : '') + date.getUTCDate(),
    ', ',
    (date.getUTCHours() < 10 ? '0' : '') + date.getUTCHours(),
    ':',
    (date.getUTCMinutes() < 10 ? '0' : '') + date.getUTCMinutes()
  ].join('')
}

function baseStatsCtrl($scope) {
  $scope.kind = 'points'
  $scope.period = 'week'
  $scope.datasetOverride = [{
    fill: false
  }]
  var update = function() {
    var labels = []
    var data = []
    var array = $scope[$scope.kind + $scope.period]
    for (var i = 0; i < array.length; ++i) {
      var d = new Date(array[i][0] * 1000)
      labels.push(prettyDate(d))
      data.push(array[i][1])
    }
    $scope.data = [data]
    $scope.labels = labels
  }
  $scope.$watchCollection('kind', update)
  $scope.$watchCollection('period', update)
}

angular.
  module('app', ['chart.js', 'ui.bootstrap']).
  controller('PointsCtrl', function($scope) {
    baseStatsCtrl($scope)
  }).
  controller('RankingCtrl', function($scope) {
    $scope.options = {
      scales: {
        yAxes: [{
          ticks: {
            reverse: true
          }
        }]
      }
    }
    baseStatsCtrl($scope)
  })
