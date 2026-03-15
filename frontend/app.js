async function loadTodos() {

 const res = await fetch("/api/todos")
 const todos = await res.json()

 const list = document.getElementById("list")
 list.innerHTML = ""

 todos.forEach(t => {
   const li = document.createElement("li")
   li.innerText = t.title
   list.appendChild(li)
 })
}

async function addTodo(){

 const input = document.getElementById("todoInput")

 await fetch("/api/create",{
   method:"POST",
   body:JSON.stringify({
      title:input.value
   })
 })

 input.value=""
 loadTodos()
}

loadTodos()