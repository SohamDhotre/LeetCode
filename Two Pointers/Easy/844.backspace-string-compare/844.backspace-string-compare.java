class Solution {
    public boolean backspaceCompare(String s, String t) {
        Stack<Character>stackS=new Stack<>();
        Stack<Character>stackT=new Stack<>();
        for(char ch:s.toCharArray()){
            if(ch=='#' && !stackS.isEmpty()) stackS.pop();
            else if(ch!='#') stackS.push(ch);
        }
        for(char ch:t.toCharArray()){
            if(ch=='#' && !stackT.isEmpty()) stackT.pop();
            else if(ch!='#') stackT.push(ch);
        }
        return stackT.equals(stackS);
    }
}