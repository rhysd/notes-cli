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
      sources = Dir["#{parent}/*.go"].reject{|p| p.end_with? '_test.go'}.join(' ')
      run "go test -v #{m[0]} #{sources} common_test.go"
      run "golint #{m[0]}"
    else
      run 'go build ./cmd/notes'
      run "golint #{m[0]}"
    end
  end
end
