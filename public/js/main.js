// nowPlaying is a global used by the global media player for track and state
// information
var nowPlaying = {
        "isPlaying": false,
        "hasPlayed": false,
        "ID": undefined,
        "path": undefined,
};

// Pause/play (pp()): Changes the pause/play icon to reflect the state of the 
// track
var track
function pp(trackID, trackPath) {
        if (nowPlaying.ID != trackID && trackID != undefined) {
                if (nowPlaying.ID != undefined) {
                track.pause();
                track.src = "../public/assets/audio/" + trackPath;
                }
                console.log("1");
                track = document.getElementById("audioID_global");
                track.addEventListener('timeupdate', (event) => {
                                tt()
                });
        }

        if (trackID == undefined && trackPath == undefined) {
                console.log("2");
                trackID = nowPlaying.ID;
        } else if (nowPlaying.path != trackPath){
                console.log("3");
                track.src = "../public/assets/audio/" + trackPath;
                nowPlaying.path = trackPath;
        }
        console.log("4");

        console.log(track, track.paused)
        if (track.paused) {
                console.log("isPaused");
                track.play();
                nowPlaying.ID = trackID;
                document.getElementById("ppImg_" + trackID).src = "public/assets/pause.png";
                document.getElementById("ppImg_global").style.backgroundImage = "url('public/assets/pause.png')";
        } else {
                console.log("isPlaying");
                track.pause();
                document.getElementById("ppImg_" + trackID).src = "public/assets/images/play.png";
                document.getElementById("ppImg_global").style.backgroundImage = "url('public/assets/images/play.png')";
        }
}

// seek() gets the mouses x-cooridinate when it clicks the outerSeeker div and
// uses this information to seek to a relative position in the audio track
function seek(e) {
        var sizer = document.getElementById("outerSeeker");
        var seekTo = ((track.duration / 100) * ((e.clientX - sizer.offsetLeft - (window.innerWidth - sizer.offsetLeft - sizer.offsetWidth)) / sizer.offsetWidth) * 100);
        track.currentTime = seekTo;
}

// Time tracker (tt()): Runs ontimeupdate and expands the innerSeeker element on
// the global player to reflect the time position of the audio track
function tt() {
        document.getElementById('innerSeeker').style.width = (Math.floor(track.currentTime) /
                Math.floor(track.duration)) * 100 + "%";
        document.getElementById("spinny").style.transform = "rotate("+Math.floor(track.currentTime)*5+"deg)";
}

// Listens for when a user chooses a new song and changes all icons in the music
// list to a play button except for the icon associated with the chosen song.
// Also adds nowPlaying information to the global "nowPlaying" object, and
// updates the info in the global player.
document.addEventListener('play', function(e) {
                document.getElementById("controls").style.display = "unset";
                if (nowPlaying.hasPlayed == false) {
                        var c1 = document.getElementById("ppImg_global").offsetWidth;
                        var c2 = document.getElementById("globalNextButt").offsetWidth;
                        var c3 = document.getElementById("globalLikeButt").offsetWidth;
                        document.getElementById("outerSeeker").style.width = (window.innerWidth - (sb+c1+c2+c3)) + "px";
                        document.getElementById("outerSeeker").style.marginLeft= (c1+c2) + "px";
                        nowPlaying.hasPlayed = true;
                } 
                updateTrackList();
}, true);

function updateTrackList() {
    var audios = document.getElementsByClassName('musicLi');
    for (var i = 0, len = audios.length; i < len; i++) {
        var pp = document.getElementById("ppImg_" + audios[i].id);
        if (audios[i].id != nowPlaying.ID) {
                // var pp = document.getElementById("ppImg_" + audios[i].id.split("_").pop());
                pp.src = "public/assets/images/play.png";
        } else {
                console.log(audios[i].id, nowPlaying.ID);
                nowPlaying.isPlaying = true;
                nowPlaying.artist = audios[i].dataset.artist;
                nowPlaying.title = audios[i].dataset.title;
                document.getElementById("globalTrackInfo").innerHTML = nowPlaying.artist +
                                                     " - " + nowPlaying.title;
                pp.src = "public/assets/images/pause.png";
        }
    }
}

function showLogin(isLoggedIn) {
        if (isLoggedIn) {
                auth("logout"); 
        } else {
                document.getElementById("hiddenAuth").style.display = "unset"; 
        }
}

function hideLogin() {
        document.getElementById("hiddenAuth").style.display = "none"; 
}

// auth is used for signing up and signing in/out. path could be:
// /api/signup
// /api/signin
// /api/logout
function auth(path) {
        var xhr = new XMLHttpRequest();

        xhr.open("POST", "/api/" + path);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.onload = function() {
                if (xhr.status === 200) {
                        var res = JSON.parse(xhr.responseText);
                        if (res.success == "false") {
                                // If we aren't successful we display an error.
                                document.getElementById("errorField").innerHTML = res.error;
                        } else {
                                // Reload the page now that the user is signed in.
                                window.location.reload();
                        }
                }
        };

        // For now, all we're sending is a username and password, but we may start
        // asking for email or mobile number at some point.
        xhr.send(JSON.stringify({
                                password: document.getElementById("password").value,
                                username: document.getElementById("username").value,
        }));
}

function like(trackID, isLoggedIn) {
        if (isLoggedIn == "false") {
                showLogin();
        } else {
                var xhr = new XMLHttpRequest();

                xhr.open("POST", "/api/like");
                xhr.setRequestHeader("Content-Type", "application/json");
                xhr.onload = function() {
                        if (xhr.status === 200) {
                                var res = JSON.parse(xhr.responseText);
                                if (res.success == "false") {
                                        // If we aren't successful we display an error.
                                        document.getElementById("errorField").innerHTML = res.error;
                                } else if (res.isLiked == "true") {
                                        document.getElementById("heart_" + trackID).style.backgroundImage = "url(/public/assets/heart_red.svg)";
                                } else if (res.isLiked == "false") {
                                        document.getElementById("heart_" + trackID).style.backgroundImage = "url(/public/assets/heart_black.svg)";
                                } else {
                                        // handle error
                                }
                        }
                };

                // For now, all we're sending is a username and password, but we may start
                // asking for email or mobile number at some point.
                xhr.send(JSON.stringify({
                                        id: trackID,
                }));

        }
}

function loadTracks(category) {
  var xhr = new XMLHttpRequest();

          xhr.open("POST", "/api/getTracks");
          xhr.setRequestHeader("Content-Type", "application/json");
          xhr.onload = function() {
                  if (xhr.status === 200) {
                          var res = JSON.parse(xhr.responseText);
                          if (res.success == "true") {
                                  var listDiv = document.getElementById("sizer");
                                  listDiv.innerHTML = res.template;
                                  updateTrackList();
                                  if (category == "FRESH" || category == "HOT") {
                                        window.history.pushState({},
                                                "page", "/#/" + category);
                                  } else {
                                        window.history.pushState({},
                                                "page", "/♥/" + category);
                                  }
                                  window.scrollTo(0, 0);
                                  // listDiv.insertAdjacentHTML("beforeend", res.template);
                          } else {
                                  // handle error
                          }
                  }
          };

          // For now, all we're sending is a username and password, but we may start
          // asking for email or mobile number at some point.
          console.log(category);
          xhr.send(JSON.stringify({
                                  category: category,
          }));

}
