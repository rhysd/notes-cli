def run(cmdline)
  puts "+#{cmdline}"
  system cmdline
end

guard :shell do
  watch /\.go$/ do |m|
    puts "#{Time.now}: #{m[0]}"
    case m[0]
    when /_test\.go$/
      parent = File.dirname m[0]
      sources = Dir["#{parent}/*.go"].reject{|p| p.end_with? '_test.go'}
      sources << m[0] << "common_test.go"
      run "go test -v #{sources.uniq.join ' '}"
      run "golint #{m[0]}"
    else
      run 'go build ./cmd/notes'
      run "golint #{m[0]}"
    end
  end
end
