guard :shell do
  watch /\.go$/ do |m|
    puts "#{Time.now}: #{m[0]}"
    case m[0]
    when /_test\.go$/
      parent = File.dirname m[0]
      sources = Dir["#{parent}/*.go"].reject{|p| p.end_with? '_test.go'}.join(' ')
      system "go test -v #{m[0]} #{sources}"
    else
      system 'go build ./cmd/notes'
    end
  end
end
