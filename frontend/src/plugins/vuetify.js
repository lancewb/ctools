import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import { md3 } from 'vuetify/blueprints'

export default createVuetify({
    blueprint: md3,
    theme: {
        defaultTheme: 'myBlueTheme',
        themes: {
            myBlueTheme: {
                dark: false,
                colors: {
                    primary: '#2196F3',
                    secondary: '#03A9F4',
                    background: '#F5F7FA',
                    surface: '#FFFFFF',
                },
            },
        },
    },
})