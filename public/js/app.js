//const { createApp } = Vue
//createApp({}).mount('#app')

//createApp
app = new Vue({
    el : '#app',
    created: function(){
        this.loadShedules()
        console.log('loadSchedules')
    },
    data() {
        return {
            message: 'Hello Vue!',
            api : '/api/jobs',
            edit: true,
            schedule : {
                ID: 0,
                name: '',
                service_name: '',
                schedule: ''
            },
            schedules: [],
            error: {
                name:'',
                name_class : '',
                service_name: '',
                service_name_class: '',
                schedule: '',
                schedule_class: ''
            },
            info: {
                edit: 'Muestra el Formulario de Edición de la Tarea seleccionada.',
                reset: 'Enviar una instruccion de Reinicio del Servicio Seleccioando, si esta en ejecución Reincia, si esta Detenido Inicia el Servicio',
                stop: 'Envia una instruccion de detener el Servicio Seleccionado, si ya esta detenido retorna que ya esta detenido!'
            }
        }
    },
    methods: {
        loadShedules: function(){
            console.log(this.api)
            axios.get(this.api).then(response=>{
                this.schedules = response.data
                console.log(this.schedules)
            }).catch(error=>{
                this.showError(error)
            })
        },
        editSchedule: function(schedule){
            console.log(schedule)
            this.schedule = schedule
            $('#create').modal('show')
            $('#name').focus()
            
            this.validateData(this.schedule.name, 'name')
            this.validateData(this.schedule.service_name, 'service_name')
            this.validateData(this.schedule.schedule, 'schedule')
            console.log('Editing..')
        },
        showError: function(error){
            console.error(error)
        },
        showCreate: function(){
            this.schedule.ID   = 0
            this.schedule.name = ''
            this.schedule.service_name = ''
            this.schedule.schedule = ''
            
            this.error.schedule_class = ''
            this.error.name_class = ''
            this.error.service_name_class = ''

            $('#create').modal('show')
            $('#name').focus()
        },
        validateData: function(data, attribute){
            console.log(data, attribute, this.error)
            switch (attribute) {
                case 'name':
                    if(!(data).trim()){
                        this.error.name = 'Debe ingresar Nombre de Tarea!'
                        this.error.name_class = 'is-invalid'
                        // /^[a-zA-ZÀ-ÿ\u00f1\u00d1]+(\s*[a-zA-ZÀ-ÿ\u00f1\u00d1]*)*[a-zA-ZÀ-ÿ\u00f1\u00d1]+$/ 
                    }
                    /*else if(data.search(this.regex)){ // /^[0-9a-zA-Z\s]*$/
                        this.error.descripcion = 'El campo debe contener solo letras o letras y numeros'
                        this.error.class = 'is-invalid' 
                    }*/else if((data).trim() && ((data).trim()).length<5 ){
                        data = data.trim()
                        this.error.name = 'El nombre de la Tarea debe tener por lo menos 5 letras.' 
                        this.error.name_class = 'is-invalid'
                    }else{
                        this.error.name = ''
                        this.error.name_class = 'is-valid'
                    }
                break
                case 'service_name':
                    if(!(data).trim()){
                        this.error.service_name = 'Debe ingresar Nombre de Servicio!'
                        this.error.service_name_class = 'is-invalid'
                        // /^[a-zA-ZÀ-ÿ\u00f1\u00d1]+(\s*[a-zA-ZÀ-ÿ\u00f1\u00d1]*)*[a-zA-ZÀ-ÿ\u00f1\u00d1]+$/ 
                    }
                    /*else if(data.search(this.regex)){ // /^[0-9a-zA-Z\s]*$/
                        this.error.descripcion = 'El campo debe contener solo letras o letras y numeros'
                        this.error.class = 'is-invalid' 
                    }*/else if((data).trim() && ((data).trim()).length<5 ){
                        data = data.trim()
                        this.error.service_name = 'El nombre del Servicio debe tener por lo menos 5 letras.' 
                        this.error.service_name_class = 'is-invalid'
                    }else{
                        this.error.service_name = ''
                        this.error.service_name_class = 'is-valid'
                    }
                break
                case 'schedule':
                    if(!(data).trim()){
                        this.error.schedule = 'Debe ingresar Horario de Programacion!'
                        this.error.schedule_class = 'is-invalid'
                    }
                    else if(data.search(/^([01]\d|2[0-3]):?([0-5]\d)$/)){ // /^[0-9a-zA-Z\s]*$/
                        this.error.schedule = 'El Horario de Programacion debe tener el formato: HH:mm'
                        this.error.schedule_class = 'is-invalid' 
                    }else if((data).trim() && ((data).trim()).length<5 ){
                        data = data.trim()
                        this.error.schedule = 'El Horario de Programacion debe tener 5 letras. HH:mm' 
                        this.error.schedule_class = 'is-invalid'
                    }else{
                        this.error.schedule = ''
                        this.error.schedule_class = 'is-valid'
                    }
                break
                default : console.log(data)
            }
            
        },
        isValid : function(){
            console.log(this.error.name_class =='is-valid' && this.error.service_name_class =='is-valid' && this.error.schedule_class =='is-valid')
            return this.error.name_class =='is-valid' && this.error.service_name_class =='is-valid' && this.error.schedule_class =='is-valid'
        },
        createSchedule: function(){
            if (this.schedule.ID > 0){
                console.log('Update...')
                axios.put('/api/job/'+this.schedule.ID, {
                    name: this.schedule.name,
                    service_name: this.schedule.service_name,
                    schedule: this.schedule.schedule
                }).then(response=> {
                    this.loadShedules()
                    $('#create').modal('hide')
                    toastr.success('Schedule Updated Successfully!', this.schedule.name)
                }).catch(error => {
                    console.error(error)
                    toastr.error(error, this.schedule.name)
                })
            }else {
                console.log('axios post')
                axios.post('/api/job', {
                    name: this.schedule.name,
                    service_name: this.schedule.service_name,
                    schedule: this.schedule.schedule
                }).then(response => {
                    console.log(response)
                    this.loadShedules()
                    $('#create').modal('hide')
                    toastr.success('Schedule Created Successfully!', this.schedule.name)
                }).catch(error => {
                    console.error(error)
                    toastr.error(error, this.schedule.name)
                })
            }
            
        },
        resetNow: function(schedule){
            axios.patch('/api/job/'+schedule.ID, {
                name: schedule.name,
                service_name: schedule.service_name,
                schedule: schedule.schedule
            }).then(response => {
                console.log(response)
                toastr.success('Schedule Restarted Successfully!', schedule.name)
            }).catch(error => {
                console.error(error)
                toastr.error(error, schedule.name)
            })
        },
        stop: function(schedule) {
            axios.patch('/api/job/stop/'+schedule.ID, {
                name: schedule.name,
                service_name: schedule.service_name,
                schedule: schedule.schedule
            }).then(response => {
                console.log(response)
                toastr.success('Schedule Stopped Successfully!', schedule.name)
            }).catch(error => {
                console.error(error)
                toastr.error(error, schedule.name)
            })
        },
        showInfo: function(message) {
            toastr.info(message, 'Info!')
        }
    }
    
}) //.mount('#app')