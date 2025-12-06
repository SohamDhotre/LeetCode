class Solution {
    public boolean backspaceCompare(String s, String t) {
        StringBuilder a=new StringBuilder(s.length());
        StringBuilder b=new StringBuilder(t.length());
        for(char p:s.toCharArray()){
            if(p=='#'){
                if(a.length()>0){
                    a.deleteCharAt(a.length()-1);
                }
            }
            else{
                a.append(p);
            }
            
        }
        for(char q:t.toCharArray()){
            if(q=='#'){
                if(b.length()>0){
                    b.deleteCharAt(b.length()-1);
                }
            }
            else{
                b.append(q);
            }
        }
        
        return a.toString().equals(b.toString());
        
    }
}