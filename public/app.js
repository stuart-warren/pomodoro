window.onload = function () {
    var display = document.querySelector('#time'),
        taskList = document.querySelector('#task-list'),
        alarm = document.querySelector('#alarm'),
        mins25 = 25*60,
        // mins25 = 5,
        timer = new CountDownTimer(mins25),
        timeObj = CountDownTimer.parse(mins25),
        eventsUrl = window.location.origin + '/events';

    format(timeObj.minutes, timeObj.seconds);

    timer.onTick(format).onTick(ifExpired);

    document.querySelector('#start-button').addEventListener('click', function () {
        timer.start();
    });

    document.addEventListener("keyup", function (k) {
      if (k.code == 'Enter')
        timer.start();
      if (k.code == 'Escape')
        alarm.pause();
      return false;
    });

    function ifExpired() {
        display.className = "notexpired";
        if (this.expired()) {
            display.className = "expired";
            completeTimer();
        }
    }

    function completeTimer() {
        var val = document.querySelector('#task-input').value;
        if (val.length === 0)
            val = "unknown task";
        alarm.play();
        fetch(encodeURI(eventsUrl+'?desc='+val), {method: "POST"})
        .then(res=>{
          getTaskLists();
        })
        .catch(error=>console.log(error));
    }

    function getTaskLists() {
      fetch(eventsUrl)
      .then(resp => resp.json())
      .then(tasks => {
        taskList.innerHTML = "";
        tasks.forEach(function(task){
          var taskContent = document.createTextNode(task.ts + ' | ' + task.desc);
          var task = document.createElement('li')
          task.appendChild(taskContent);
          taskList.insertBefore(task, taskList.firstElementChild);
        });
      })
      .catch(error=>console.log(error));
    }

    function format(minutes, seconds) {
        minutes = minutes < 10 ? "0" + minutes : minutes;
        seconds = seconds < 10 ? "0" + seconds : seconds;
        display.textContent = minutes + 'm ' + seconds + 's';
    }

    getTaskLists();
};
