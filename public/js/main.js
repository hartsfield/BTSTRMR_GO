// nowPlaying is a global used by the global media player for track and state
// information
var nowPlaying = {
    "isPlaying": false
};

// Pause/play (pp()): Changes the pause/play icon to reflect the state of the 
// track
function pp(trackID) {
    if (trackID == undefined) {
        trackID = nowPlaying.ID;
    }
    var track = document.getElementById("audioID_" + trackID);
    if (track.paused) {
        document.getElementById("ppImg_" + trackID).src = "public/assets/images/pause.png";
        document.getElementById("globalPPImg").src = "public/assets/images/pause.png";
        nowPlaying.ID = trackID;
        track.play();
    } else {
        track.pause();
        document.getElementById("ppImg_" + trackID).src = "public/assets/images/play.png";
        document.getElementById("globalPPImg").src = "public/assets/images/play.png";
    }
}

// seek() gets the mouses x-cooridinate when it clicks the outerSeeker div and
// uses this information to seek to a relative position in the audio track
function seek(e) {
    var track = document.getElementById("audioID_" + nowPlaying.ID);
    var seekTo = ((track.duration / 100) * (e.clientX / window.innerWidth) * 100);
    track.currentTime = seekTo;
}

// Time tracker (tt()): Runs ontimeupdate and expands the innerSeeker element on
// the global player to reflect the time position of the audio track
function tt(trackID) {
    track = document.getElementById("audioID_" + trackID);
    document.getElementById('innerSeeker').style.width = (Math.floor(track.currentTime) /
        Math.floor(track.duration)) * 100 + "%";
}

// Listens for when a user chooses a new song and changes all icons in the music
// list to a play button except for the icon associated with the chosen song.
// Also adds nowPlaying information to the global "nowPlaying" object, and
// updates the info in the global player.
document.addEventListener('play', function(e) {
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
