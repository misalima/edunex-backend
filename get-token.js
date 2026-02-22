import { createClient } from '@supabase/supabase-js'

const supabaseUrl = 'https://kkpfbmnndjsvwnwaquaf.supabase.co'
const supabaseAnonKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtrcGZibW5uZGpzdndud2FxdWFmIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzE1MjQzNjIsImV4cCI6MjA4NzEwMDM2Mn0.Q9E4ZrNG3Ur_Bs7BWPrzD19ATzAqMhpLq86Z1ok90bc'

const supabase = createClient(supabaseUrl, supabaseAnonKey)

async function main() {
    const { data, error } = await supabase.auth.signInWithPassword({
        email: 'misael.lima@professor.educ.al.gov.br',
        password: 'Admin123',
    })

    if (error) {
        console.error('Erro no login:', error.message)
        process.exit(1)
    }

    console.log('Access Token:', data.session.access_token)
}

main()