// nowPlaying is a global used by the global media player for track and state
// information
var nowPlaying = {
        "isPlaying": false,
        "hasPlayed": false
};

// Pause/play (pp()): Changes the pause/play icon to reflect the state of the 
// track
function pp(trackID) {
        if (trackID == undefined) {
                trackID = nowPlaying.ID;
        }
        var track = document.getElementById("audioID_" + trackID);
        if (track.paused) {
                document.getElementById("ppImg_" + trackID).src = "public/assets/pause.png";
                document.getElementById("globalPPButt").style.backgroundImage = "url('public/assets/pause.png')";
                nowPlaying.ID = trackID;
                track.play();
        } else {
                track.pause();
                document.getElementById("ppImg_" + trackID).src = "public/assets/images/play.png";
                document.getElementById("globalPPButt").style.backgroundImage = "url('public/assets/images/play.png')";
        }
}

// seek() gets the mouses x-cooridinate when it clicks the outerSeeker div and
// uses this information to seek to a relative position in the audio track
function seek(e) {
        var track = document.getElementById("audioID_" + nowPlaying.ID);
        var sizer = document.getElementById("outerSeeker")
        console.log(e.clientX);
        var seekTo = ((track.duration / 100) * ((e.clientX - sizer.offsetLeft - (window.innerWidth - sizer.offsetLeft - sizer.offsetWidth)) / sizer.offsetWidth) * 100);
        track.currentTime = seekTo;
}

// Time tracker (tt()): Runs ontimeupdate and expands the innerSeeker element on
// the global player to reflect the time position of the audio track
function tt(trackID) {
        track = document.getElementById("audioID_" + trackID);
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
                        var c1 = document.getElementById("globalPPButt").offsetWidth;
                        var c2 = document.getElementById("globalNextButt").offsetWidth;
                        var c3 = document.getElementById("globalLikeButt").offsetWidth;
                        document.getElementById("outerSeeker").style.width = (window.innerWidth - (sb+c1+c2+c3)) + "px";
                        document.getElementById("outerSeeker").style.marginLeft= (c1+c2) + "px";
                        nowPlaying.hasPlayed = true;
                } 


                var audios = document.getElementsByTagName('audio');
                for (var i = 0, len = audios.length; i < len; i++) {
                        if (audios[i] != e.target) {
                                audios[i].pause();
                                var pp = document.getElementById("ppImg_" + audios[i].id.split("_").pop());
                                pp.src = "public/assets/images/play.png";
                        } else {
                                nowPlaying.artist = audios[i].dataset.artist;
                                nowPlaying.title = audios[i].dataset.title;
                                nowPlaying.isPlaying = true;
                                document.getElementById("globalTrackInfo").innerHTML = nowPlaying.artist +
                                                                     " - " + nowPlaying.title;
                        }
                }
}, true);

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
