function updateUsers(jsondata){
    if(jsondata.id == ""){
        return
    }
    var user = document.getElementById("user_"+jsondata.id);
    if(user == null){

        var name = jsondata.first_name;

        var li = document.createElement('li');
        li.className = 'bg-base-100 rounded p-2 my-2';
        li.id = 'user_' + name;

        var span = document.createElement('span');
        span.className = 'indicator-item badge badge-xs badge-success';
        span.id = 'status_' + name;
    
        li.appendChild(span);
    
        var textNode = document.createTextNode(name);
        li.appendChild(textNode);
    
        var userList = document.getElementById('users_list');
        userList.appendChild(li);
    }

    // var status = document.getElementById("status_"+jsondata.name);
    // if(jsondata.IsOnline){
    //     status?.classList.add("badge-success");
    // }else{
    //     status?.classList.remove("badge-success");
    // }
}